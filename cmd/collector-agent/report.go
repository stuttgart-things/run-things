/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Reporter pushes inventory snapshots and heartbeats to the run-things server.
type Reporter struct {
	Cfg        Config
	Discoverer *Discoverer
	client     *http.Client
}

func (r *Reporter) http() *http.Client {
	if r.client == nil {
		r.client = &http.Client{Timeout: r.Cfg.HTTPTimeout}
	}
	return r.client
}

// RunReportLoop sends a full inventory snapshot every ReportInterval. It also
// sends an immediate snapshot at startup so the server has data right away.
func (r *Reporter) RunReportLoop(ctx context.Context) {
	r.reportOnce(ctx)
	t := time.NewTicker(r.Cfg.ReportInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.reportOnce(ctx)
		}
	}
}

// RunHeartbeatLoop sends a lightweight "I'm alive" ping every HeartbeatInterval
// so the server's last-seen timestamp stays fresh between full reports.
func (r *Reporter) RunHeartbeatLoop(ctx context.Context) {
	t := time.NewTicker(r.Cfg.HeartbeatInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.heartbeatOnce(ctx)
		}
	}
}

func (r *Reporter) reportOnce(ctx context.Context) {
	inv, err := r.Discoverer.Snapshot(ctx)
	if err != nil {
		log.Printf("snapshot failed: %v", err)
		return
	}
	if err := r.post(ctx, "/api/v1/collector/inventory", inv); err != nil {
		log.Printf("inventory report failed: %v", err)
		return
	}
	log.Printf("REPORTED INVENTORY: deployments=%d statefulsets=%d daemonsets=%d services=%d ingresses=%d",
		len(inv.Deployments), len(inv.StatefulSets), len(inv.DaemonSets),
		len(inv.Services), len(inv.Ingresses))
}

func (r *Reporter) heartbeatOnce(ctx context.Context) {
	hb := Heartbeat{ClusterName: r.Cfg.ClusterName, Endpoint: r.Cfg.Endpoint}
	if err := r.post(ctx, "/api/v1/collector/heartbeat", hb); err != nil {
		log.Printf("heartbeat failed: %v", err)
	}
}

func (r *Reporter) post(ctx context.Context, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Cfg.ServerURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if r.Cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.Cfg.AuthToken)
	}
	resp, err := r.http().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
