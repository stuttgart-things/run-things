/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"log"
	"sync"
	"time"
)

// ClusterStore holds inventory data from all connected collectors
type ClusterStore struct {
	mu         sync.RWMutex
	clusters   map[string]*ClusterInventory
	clusterMeta map[string]*ClusterInfo
}

// NewClusterStore creates a new cluster store
func NewClusterStore() *ClusterStore {
	return &ClusterStore{
		clusters:    make(map[string]*ClusterInventory),
		clusterMeta: make(map[string]*ClusterInfo),
	}
}

// UpdateInventory updates the inventory for a cluster
func (cs *ClusterStore) UpdateInventory(inv *ClusterInventory) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	inv.LastUpdated = time.Now()
	cs.clusters[inv.ClusterName] = inv
	cs.clusterMeta[inv.ClusterName] = &ClusterInfo{
		ClusterName: inv.ClusterName,
		LastSeen:    time.Now(),
	}
	log.Printf("UPDATED INVENTORY FOR CLUSTER %s: %d deployments, %d statefulsets, %d daemonsets",
		inv.ClusterName, len(inv.Deployments), len(inv.StatefulSets), len(inv.DaemonSets))
}

// Heartbeat updates the last seen time for a cluster
func (cs *ClusterStore) Heartbeat(clusterName, endpoint string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if meta, ok := cs.clusterMeta[clusterName]; ok {
		meta.LastSeen = time.Now()
		meta.Endpoint = endpoint
	} else {
		cs.clusterMeta[clusterName] = &ClusterInfo{
			ClusterName: clusterName,
			Endpoint:    endpoint,
			LastSeen:    time.Now(),
		}
	}
}

// GetAllClusters returns all cluster info
func (cs *ClusterStore) GetAllClusters() []ClusterInfo {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	result := make([]ClusterInfo, 0, len(cs.clusterMeta))
	for _, info := range cs.clusterMeta {
		result = append(result, *info)
	}
	return result
}

// GetInventory returns the inventory for a specific cluster
func (cs *ClusterStore) GetInventory(clusterName string) (*ClusterInventory, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	inv, ok := cs.clusters[clusterName]
	return inv, ok
}

// GetAllInventories returns all cluster inventories
func (cs *ClusterStore) GetAllInventories() map[string]*ClusterInventory {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	out := make(map[string]*ClusterInventory, len(cs.clusters))
	for k, v := range cs.clusters {
		out[k] = v
	}
	return out
}
