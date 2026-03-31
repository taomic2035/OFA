package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/service"
	"github.com/ofa/center/pkg/grpc"
	"github.com/ofa/center/pkg/rest"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/center.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	centerService, err := service.NewCenterService(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create center service: %v", err)
	}

	// Start gRPC server
	grpcServer := grpc.NewServer(centerService)
	go func() {
		if err := grpcServer.Start(cfg.GRPC.Address); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start REST API server
	restServer := rest.NewServer(centerService, cfg)
	go func() {
		if err := restServer.Start(cfg.REST.Address); err != nil {
			log.Fatalf("Failed to start REST server: %v", err)
		}
	}()

	log.Printf("OFA Center started - gRPC: %s, REST: %s", cfg.GRPC.Address, cfg.REST.Address)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down Center...")

	// Graceful shutdown
	grpcServer.Stop()
	restServer.Stop()
	centerService.Close()

	log.Println("Center stopped")
}