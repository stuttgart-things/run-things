/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"log"
	"os"
	"strings"
	"time"
)

// Config holds the collector agent's runtime configuration.
type Config struct {
	ClusterName       string
	ServerURL         string
	AuthToken         string
	Endpoint          string
	Kubeconfig        string
	Namespaces        []string
	ReportInterval    time.Duration
	HeartbeatInterval time.Duration
	HTTPTimeout       time.Duration
}

// loadConfig reads agent configuration from environment variables. CLUSTER_NAME
// and SERVER_URL are required. Sensible defaults apply to everything else.
func loadConfig() Config {
	cfg := Config{
		ClusterName:       os.Getenv("CLUSTER_NAME"),
		ServerURL:         strings.TrimRight(os.Getenv("SERVER_URL"), "/"),
		AuthToken:         os.Getenv("COLLECTOR_TOKEN"),
		Endpoint:          os.Getenv("CLUSTER_ENDPOINT"),
		Kubeconfig:        os.Getenv("KUBECONFIG"),
		Namespaces:        splitCSV(os.Getenv("NAMESPACES")),
		ReportInterval:    parseDuration("REPORT_INTERVAL", 60*time.Second),
		HeartbeatInterval: parseDuration("HEARTBEAT_INTERVAL", 30*time.Second),
		HTTPTimeout:       parseDuration("HTTP_TIMEOUT", 15*time.Second),
	}

	if cfg.ClusterName == "" {
		log.Fatalf("CLUSTER_NAME is required")
	}
	if cfg.ServerURL == "" {
		log.Fatalf("SERVER_URL is required (e.g. https://run-things.example.com)")
	}
	return cfg
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDuration(envKey string, def time.Duration) time.Duration {
	v := os.Getenv(envKey)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("invalid %s=%q, using default %s", envKey, v, def)
		return def
	}
	return d
}
