/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// CollectorAuthToken is the optional bearer token expected from collector
// agents reporting in. When empty, authentication is disabled.
var CollectorAuthToken = os.Getenv("COLLECTOR_TOKEN")

// CollectorHeartbeat is the JSON payload accepted on the heartbeat endpoint.
type CollectorHeartbeat struct {
	ClusterName string `json:"clusterName"`
	Endpoint    string `json:"endpoint,omitempty"`
}

// RegisterCollectorRoutes wires the collector ingest endpoints onto the given
// mux. Agents running inside member clusters POST inventory snapshots and
// heartbeats here so the central run-things server can keep its inventory
// fresh.
func RegisterCollectorRoutes(mux *http.ServeMux, cs *ClusterStore) {
	mux.HandleFunc("POST /api/v1/collector/inventory", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeCollector(w, r) {
			return
		}
		var inv ClusterInventory
		if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
			http.Error(w, `{"error":"invalid inventory payload"}`, http.StatusBadRequest)
			return
		}
		if inv.ClusterName == "" {
			http.Error(w, `{"error":"clusterName is required"}`, http.StatusBadRequest)
			return
		}
		cs.UpdateInventory(&inv)
		writeCollectorAck(w, "inventory accepted")
	})

	mux.HandleFunc("POST /api/v1/collector/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeCollector(w, r) {
			return
		}
		var hb CollectorHeartbeat
		if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
			http.Error(w, `{"error":"invalid heartbeat payload"}`, http.StatusBadRequest)
			return
		}
		if hb.ClusterName == "" {
			http.Error(w, `{"error":"clusterName is required"}`, http.StatusBadRequest)
			return
		}
		cs.Heartbeat(hb.ClusterName, hb.Endpoint)
		log.Printf("HEARTBEAT FROM CLUSTER %s endpoint=%s", hb.ClusterName, hb.Endpoint)
		writeCollectorAck(w, "heartbeat accepted")
	})
}

func authorizeCollector(w http.ResponseWriter, r *http.Request) bool {
	if CollectorAuthToken == "" {
		return true
	}
	if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == CollectorAuthToken {
		return true
	}
	http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	return false
}

func writeCollectorAck(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "message": msg})
}
