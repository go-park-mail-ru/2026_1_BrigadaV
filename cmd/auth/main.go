package main

import (
	"log"
	"net"

	"guidely-app/internal/auth"
	"guidely-app/internal/auth/repository"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	pb "guidely-app/pkg/pb/auth"

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

	adapter := &repository.PgxPoolAdapter{Pool: pool}
	userRepo := repository.NewUserRepo(adapter)
	sessRepo := repository.NewSessionRepo(adapter)

	svc := auth.NewService(userRepo, sessRepo)
	server := auth.NewServer(svc)

	lis, _ := net.Listen("tcp", ":8085")
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, server)
	reflection.Register(s)

	log.Println("Auth service started on :8085")
	log.Fatal(s.Serve(lis))
}
