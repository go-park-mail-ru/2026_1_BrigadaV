package main

import (
	"log"
	"net"

	"guidely-app/internal/auth"
	"guidely-app/internal/auth/repository"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	"guidely-app/pkg/metrics"
	pb "guidely-app/pkg/pb/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load error:", err)
	}
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	metrics.StartMetricsServer("9101")

	adapter := &repository.PgxPoolAdapter{Pool: pool}
	userRepo := repository.NewUserRepo(adapter)
	sessRepo := repository.NewSessionRepo(adapter)

	svc := auth.NewService(userRepo, sessRepo)
	server := auth.NewServer(svc)

	lis, err := net.Listen("tcp", ":8085")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	pb.RegisterAuthServiceServer(s, server)
	reflection.Register(s)

	log.Println("Auth service started on :8085, metrics :9101")
	log.Fatal(s.Serve(lis))
}
