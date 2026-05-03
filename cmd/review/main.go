package main

import (
	"log"
	"net"

	"guidely-app/internal/review"
	"guidely-app/internal/review/repository"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
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

	reviewRepo := repository.NewReviewRepo(pool)
	svc := review.NewService(reviewRepo)
	server := review.NewServer(svc)

	lis, _ := net.Listen("tcp", ":50053")
	s := grpc.NewServer()
	pb.RegisterReviewServiceServer(s, server)
	reflection.Register(s)

	log.Println("Review service started on :50053")
	log.Fatal(s.Serve(lis))
}
