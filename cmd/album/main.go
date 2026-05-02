package main

import (
	"log"
	"net"

	"guidely-app/internal/album"
	"guidely-app/internal/album/repository"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	pb "guidely-app/pkg/pb/album"

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

	albumRepo := repository.NewAlbumRepo(pool)
	svc := album.NewService(albumRepo)
	server := album.NewServer(svc)

	lis, _ := net.Listen("tcp", ":50052")
	s := grpc.NewServer()
	pb.RegisterAlbumServiceServer(s, server)
	reflection.Register(s)

	log.Println("Album service started on :50052")
	log.Fatal(s.Serve(lis))
}
