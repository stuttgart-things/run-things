/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"log"
	"net"
	"os"

	"github.com/pterm/pterm"
	"github.com/stuttgart-things/weRunThings/internal"

	"google.golang.org/grpc"
)

const (
	defaultGRPCPort = ":50051"
	defaultHTTPPort = "8080"
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
		serverPort = defaultGRPCPort
	} else {
		serverPort = ":" + serverPort
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

	// START HTTP/HTMX SERVER IN BACKGROUND
	go internal.StartWebServer(httpPort, monitor, clusterStore, loadConfigFrom, configLocation, configName)

	// START GRPC SERVER
	lis, err := net.Listen("tcp", serverPort)
	if err != nil {
		log.Fatalf("FAILED TO LISTEN: %v", err)
	}

	s := grpc.NewServer()
	// TODO: Register CollectorService when proto is compiled
	// collectorservice.RegisterCollectorServiceServer(s, &collectorServer{store: clusterStore})

	log.Printf("GRPC SERVER LISTENING AT %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("FAILED TO SERVE: %v", err)
	}
}
