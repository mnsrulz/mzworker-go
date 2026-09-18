package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/mnsrulz/mzworker-go/gen/options"
)

const (
	defaultListenAddr = "0.0.0.0:50051"
	defaultDataDir    = "data/options_data"
)

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("Failed to resolve data directory: %v", err)
	}

	if _, err := os.Stat(absDataDir); os.IsNotExist(err) {
		log.Fatalf("Data directory does not exist: %s", absDataDir)
	}

	log.Printf("Server starting on %s", listenAddr)
	log.Printf("Data directory: %s", absDataDir)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	svc := NewServer(absDataDir)
	pb.RegisterOptionsQueryServiceServer(s, svc)
	reflection.Register(s)

	go func() {
		fmt.Printf("Server listening on %s\n", listenAddr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	s.GracefulStop()
}
