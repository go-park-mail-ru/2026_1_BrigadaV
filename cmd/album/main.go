package main

import (
	"log"
	"net"

	"guidely-app/internal/album"
	"guidely-app/internal/album/repository"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	"guidely-app/pkg/metrics"
	pb "guidely-app/pkg/pb/album"

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

	metrics.StartMetricsServer("9102")

	albumRepo := repository.NewAlbumRepo(pool)
	svc := album.NewService(albumRepo)
	server := album.NewServer(svc)

	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	pb.RegisterAlbumServiceServer(s, server)
	reflection.Register(s)

	log.Println("Album service started on :8086, metrics :9102")
	log.Fatal(s.Serve(lis))
}
