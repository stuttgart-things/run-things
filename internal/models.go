/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"sync"
	"time"
)

// Service represents a monitored service in the portal
type Service struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Category    string            `yaml:"category" json:"category"`
	URL         string            `yaml:"url" json:"url"`
	LogoURL     string            `yaml:"logoURL,omitempty" json:"logoURL,omitempty"`
	Icon        string            `yaml:"icon,omitempty" json:"icon,omitempty"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	HealthCheck HealthCheckConfig `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
}

// HealthCheckConfig configures how health checks are performed
type HealthCheckConfig struct {
	Enabled        bool              `yaml:"enabled" json:"enabled"`
	Interval       int               `yaml:"interval" json:"interval"`             // seconds
	Method         string            `yaml:"method,omitempty" json:"method,omitempty"` // GET, POST, HEAD
	ExpectedStatus int               `yaml:"expectedStatus,omitempty" json:"expectedStatus,omitempty"`
	ExpectedBody   string            `yaml:"expectedBody,omitempty" json:"expectedBody,omitempty"`
	Headers        map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	TLSCheck       bool              `yaml:"tlsCheck,omitempty" json:"tlsCheck,omitempty"`
	Timeout        int               `yaml:"timeout,omitempty" json:"timeout,omitempty"` // seconds
	Body           string            `yaml:"body,omitempty" json:"body,omitempty"`
}

// CheckResult holds the result of a single health check
type CheckResult struct {
	Timestamp    time.Time `json:"timestamp"`
	StatusCode   int       `json:"statusCode"`
	ResponseTime int64     `json:"responseTime"` // milliseconds
	Status       string    `json:"status"`       // UP, DOWN, DEGRADED
	Error        string    `json:"error,omitempty"`
	TLSExpiry    time.Time `json:"tlsExpiry,omitempty"`
	TLSDaysLeft  int       `json:"tlsDaysLeft,omitempty"`
	BodyMatched  bool      `json:"bodyMatched"`
}

// ServiceState holds a service and its health check history
type ServiceState struct {
	Service       Service       `json:"service"`
	Results       []CheckResult `json:"results"`
	CurrentStatus string        `json:"currentStatus"` // UP, DOWN, DEGRADED, UNKNOWN
	mu            sync.RWMutex
}

// AddResult adds a check result to the ring buffer (max 60)
func (s *ServiceState) AddResult(r CheckResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Results = append(s.Results, r)
	if len(s.Results) > 60 {
		s.Results = s.Results[len(s.Results)-60:]
	}
	s.CurrentStatus = r.Status
}

// GetResults returns a copy of the results
func (s *ServiceState) GetResults() []CheckResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CheckResult, len(s.Results))
	copy(out, s.Results)
	return out
}

// GetCurrentStatus returns the current status
func (s *ServiceState) GetCurrentStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.CurrentStatus == "" {
		return "UNKNOWN"
	}
	return s.CurrentStatus
}

// GetLastResult returns the most recent result
func (s *ServiceState) GetLastResult() (CheckResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.Results) == 0 {
		return CheckResult{}, false
	}
	return s.Results[len(s.Results)-1], true
}

// ClusterInfo tracks a connected collector cluster
type ClusterInfo struct {
	ClusterName string    `json:"clusterName"`
	Endpoint    string    `json:"endpoint,omitempty"`
	LastSeen    time.Time `json:"lastSeen"`
}

// WorkloadInfo describes a single K8s workload
type WorkloadInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Kind         string            `json:"kind"`
	Replicas     int32             `json:"replicas"`
	Ready        int32             `json:"ready"`
	Images       []string          `json:"images,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreationTime time.Time         `json:"creationTime"`
}

// ClusterInventory holds the full workload inventory for a cluster
type ClusterInventory struct {
	ClusterName  string         `json:"clusterName"`
	Deployments  []WorkloadInfo `json:"deployments,omitempty"`
	StatefulSets []WorkloadInfo `json:"statefulsets,omitempty"`
	DaemonSets   []WorkloadInfo `json:"daemonsets,omitempty"`
	Services     []WorkloadInfo `json:"services,omitempty"`
	Ingresses    []WorkloadInfo `json:"ingresses,omitempty"`
	LastUpdated  time.Time      `json:"lastUpdated"`
}

// ServiceConfig is the top-level config loaded from YAML
type ServiceConfig struct {
	Services []Service `yaml:"services" json:"services"`
}
