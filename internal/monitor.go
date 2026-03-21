/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Monitor runs background health checks for all services
type Monitor struct {
	mu       sync.RWMutex
	states   map[string]*ServiceState
	client   *http.Client
	stopCh   chan struct{}
	loadFrom string
	configLoc string
	configNm  string
}

// NewMonitor creates a new health check monitor
func NewMonitor(loadFrom, configLoc, configNm string) *Monitor {
	return &Monitor{
		states: make(map[string]*ServiceState),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		stopCh:    make(chan struct{}),
		loadFrom:  loadFrom,
		configLoc: configLoc,
		configNm:  configNm,
	}
}

// LoadAndStart loads services and starts monitoring
func (m *Monitor) LoadAndStart() {
	services := LoadServices(m.loadFrom, m.configLoc, m.configNm)
	m.mu.Lock()
	for _, svc := range services {
		m.states[svc.Name] = &ServiceState{
			Service:       svc,
			CurrentStatus: "UNKNOWN",
		}
	}
	m.mu.Unlock()

	go m.run()
	log.Printf("HEALTH MONITOR STARTED FOR %d SERVICES", len(services))
}

// run is the main monitor loop
func (m *Monitor) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Run initial checks immediately
	m.checkAll()

	for {
		select {
		case <-ticker.C:
			m.checkAll()
		case <-m.stopCh:
			return
		}
	}
}

// checkAll checks all services that are due for a check
func (m *Monitor) checkAll() {
	m.mu.RLock()
	states := make([]*ServiceState, 0, len(m.states))
	for _, s := range m.states {
		states = append(states, s)
	}
	m.mu.RUnlock()

	for _, state := range states {
		if !state.Service.HealthCheck.Enabled {
			continue
		}

		// Check if enough time has passed since last check
		interval := time.Duration(state.Service.HealthCheck.Interval) * time.Second
		if interval == 0 {
			interval = 30 * time.Second
		}

		last, hasLast := state.GetLastResult()
		if hasLast && time.Since(last.Timestamp) < interval {
			continue
		}

		result := m.checkService(state.Service)
		state.AddResult(result)
	}
}

// checkService performs a single health check
func (m *Monitor) checkService(svc Service) CheckResult {
	result := CheckResult{
		Timestamp: time.Now(),
	}

	method := svc.HealthCheck.Method
	if method == "" {
		method = "GET"
	}

	timeout := time.Duration(svc.HealthCheck.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var bodyReader io.Reader
	if svc.HealthCheck.Body != "" {
		bodyReader = strings.NewReader(svc.HealthCheck.Body)
	}

	req, err := http.NewRequest(method, svc.URL, bodyReader)
	if err != nil {
		result.Status = "DOWN"
		result.Error = fmt.Sprintf("request creation failed: %v", err)
		return result
	}

	for k, v := range svc.HealthCheck.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	result.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		result.Status = "DOWN"
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Check expected status
	expectedStatus := svc.HealthCheck.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	if resp.StatusCode != expectedStatus {
		result.Status = "DOWN"
		result.Error = fmt.Sprintf("expected status %d, got %d", expectedStatus, resp.StatusCode)
		return result
	}

	// Check expected body
	if svc.HealthCheck.ExpectedBody != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Status = "DEGRADED"
			result.Error = "failed to read response body"
			return result
		}
		if strings.Contains(string(body), svc.HealthCheck.ExpectedBody) {
			result.BodyMatched = true
		} else {
			result.Status = "DEGRADED"
			result.Error = "expected body not found"
			return result
		}
	}

	// Check TLS if enabled
	if svc.HealthCheck.TLSCheck {
		tlsInfo, err := CheckTLSCertificate(svc.URL)
		if err == nil {
			result.TLSExpiry = tlsInfo.Expiry
			result.TLSDaysLeft = tlsInfo.DaysLeft
			if tlsInfo.DaysLeft < 14 {
				result.Status = "DEGRADED"
				result.Error = fmt.Sprintf("TLS cert expires in %d days", tlsInfo.DaysLeft)
				return result
			}
		}
	}

	// Check response time for degraded status (>5s)
	if result.ResponseTime > 5000 {
		result.Status = "DEGRADED"
		return result
	}

	result.Status = "UP"
	return result
}

// GetStates returns all service states (thread-safe)
func (m *Monitor) GetStates() map[string]*ServiceState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*ServiceState, len(m.states))
	for k, v := range m.states {
		out[k] = v
	}
	return out
}

// GetState returns a single service state
func (m *Monitor) GetState(name string) (*ServiceState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.states[name]
	return s, ok
}

// GetServices returns the current list of services
func (m *Monitor) GetServices() []Service {
	m.mu.RLock()
	defer m.mu.RUnlock()
	services := make([]Service, 0, len(m.states))
	for _, s := range m.states {
		services = append(services, s.Service)
	}
	return services
}

// AddService adds a service to monitoring
func (m *Monitor) AddService(svc Service) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[svc.Name] = &ServiceState{
		Service:       svc,
		CurrentStatus: "UNKNOWN",
	}
}

// RemoveService removes a service from monitoring
func (m *Monitor) RemoveService(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, name)
}

// UpdateService updates a service in monitoring
func (m *Monitor) UpdateService(svc Service) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.states[svc.Name]; ok {
		existing.Service = svc
	} else {
		m.states[svc.Name] = &ServiceState{
			Service:       svc,
			CurrentStatus: "UNKNOWN",
		}
	}
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.stopCh)
}
