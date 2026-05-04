package main

import (
	"log"
	"net"

	"guidely-app/internal/review"
	"guidely-app/internal/review/repository"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	"guidely-app/pkg/metrics"
	pb "guidely-app/pkg/pb/review"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, _ := config.Load()
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	metrics.StartMetricsServer("9103")

	reviewRepo := repository.NewReviewRepo(pool)
	svc := review.NewService(reviewRepo)
	server := review.NewServer(svc)

	lis, _ := net.Listen("tcp", ":8087")
	s := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	pb.RegisterReviewServiceServer(s, server)
	reflection.Register(s)

	log.Println("Review service started on :8087, metrics :9103")
	log.Fatal(s.Serve(lis))
}
