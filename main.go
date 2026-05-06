/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pterm/pterm"
	"github.com/stuttgart-things/run-things/internal"
)

const (
	defaultCollectorPort = "50051"
	defaultHTTPPort      = "8080"
)

var (
	logger         = pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace)
	loadConfigFrom = os.Getenv("LOAD_CONFIG_FROM")
	configName     = os.Getenv("CONFIG_NAME")
	configLocation = os.Getenv("CONFIG_LOCATION")
	serverPort     = os.Getenv("SERVER_PORT")
	httpPort       = os.Getenv("HTTP_PORT")
)

func main() {
	// PRINT BANNER + VERSION INFO
	internal.PrintBanner()

	if loadConfigFrom == "" {
		loadConfigFrom = "disk"
	}
	if configLocation == "" {
		configLocation = "tests"
	}
	if configName == "" {
		configName = "services.yaml"
	}

	if serverPort == "" {
		serverPort = defaultCollectorPort
	}

	if httpPort == "" {
		httpPort = defaultHTTPPort
	}

	// CREATE CLUSTER STORE FOR COLLECTOR DATA
	clusterStore := internal.NewClusterStore()

	// START HEALTH MONITOR
	monitor := internal.NewMonitor(loadConfigFrom, configLocation, configName)
	monitor.LoadAndStart()

	logger.Info("LOAD CONFIG FROM", logger.Args("", loadConfigFrom))
	logger.Info("CONFIG LOCATION", logger.Args("", configLocation))
	logger.Info("CONFIG NAME", logger.Args("", configName))

	// START HTTP/HTMX SERVER IN BACKGROUND (dashboard + REST API on httpPort)
	go internal.StartWebServer(httpPort, monitor, clusterStore, loadConfigFrom, configLocation, configName)

	// START COLLECTOR INGEST SERVER (cluster agents POST inventory + heartbeats here)
	collectorMux := http.NewServeMux()
	internal.RegisterCollectorRoutes(collectorMux, clusterStore)
	log.Printf("COLLECTOR INGEST SERVER LISTENING AT :%s", serverPort)
	if err := http.ListenAndServe(":"+serverPort, collectorMux); err != nil {
		log.Fatalf("FAILED TO SERVE COLLECTOR INGEST: %v", err)
	}
}
