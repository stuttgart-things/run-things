/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de

run-things collector agent

A small daemon that runs inside a Kubernetes cluster, periodically scrapes the
local cluster's workload inventory (Deployments, StatefulSets, DaemonSets,
Services, Ingresses), and pushes it to a central run-things server via REST.
It also sends a lightweight heartbeat so the server can mark the cluster as
alive even between full reports.
*/
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := loadConfig()

	log.Printf("STARTING run-things collector agent")
	log.Printf("  cluster=%s server=%s report=%s heartbeat=%s namespaces=%v",
		cfg.ClusterName, cfg.ServerURL, cfg.ReportInterval, cfg.HeartbeatInterval, cfg.Namespaces)

	dyn, err := newDynamicClient(cfg.Kubeconfig)
	if err != nil {
		log.Fatalf("FAILED TO CREATE K8S CLIENT: %v", err)
	}

	disco := &Discoverer{Client: dyn, Namespaces: cfg.Namespaces, ClusterName: cfg.ClusterName}
	rep := &Reporter{Cfg: cfg, Discoverer: disco}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go rep.RunReportLoop(ctx)
	go rep.RunHeartbeatLoop(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("SHUTTING DOWN")
	cancel()
	time.Sleep(500 * time.Millisecond)
}
