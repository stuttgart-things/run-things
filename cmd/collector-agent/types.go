/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import "time"

// WorkloadInfo mirrors the shape expected by the run-things server. Field tags
// must match internal.WorkloadInfo so the JSON payload deserializes cleanly on
// the server side.
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

// ClusterInventory mirrors internal.ClusterInventory on the server.
type ClusterInventory struct {
	ClusterName  string         `json:"clusterName"`
	Deployments  []WorkloadInfo `json:"deployments,omitempty"`
	StatefulSets []WorkloadInfo `json:"statefulsets,omitempty"`
	DaemonSets   []WorkloadInfo `json:"daemonsets,omitempty"`
	Services     []WorkloadInfo `json:"services,omitempty"`
	Ingresses    []WorkloadInfo `json:"ingresses,omitempty"`
	LastUpdated  time.Time      `json:"lastUpdated"`
}

// Heartbeat mirrors internal.CollectorHeartbeat.
type Heartbeat struct {
	ClusterName string `json:"clusterName"`
	Endpoint    string `json:"endpoint,omitempty"`
}
