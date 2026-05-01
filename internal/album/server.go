package album

import (
	"context"
	"guidely-app/pkg/models"
	pb "guidely-app/pkg/pb/album"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedAlbumServiceServer
	svc *Service
}

func NewServer(svc *Service) *Server {
	return &Server{svc: svc}
}

func toAlbum(a *models.Album) *pb.Album {
	return &pb.Album{
		Id:           a.ID,
		TripId:       a.TripID,
		Name:         a.Name,
		Description:  a.Description,
		CoverPhotoId: a.CoverPhotoID,
		MaxPhotos:    int32(a.MaxPhotos),
	}
}

func (s *Server) Create(ctx context.Context, req *pb.CreateAlbumRequest) (*pb.Album, error) {
	a := &models.Album{TripID: req.TripId, Name: req.Name, Description: req.Description, MaxPhotos: int(req.MaxPhotos)}
	if err := s.svc.Create(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toAlbum(a), nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetAlbumRequest) (*pb.Album, error) {
	a, err := s.svc.GetByID(ctx, req.Id)
	if err != nil || a == nil {
		return nil, status.Error(codes.NotFound, "album not found")
	}
	return toAlbum(a), nil
}

func (s *Server) Update(ctx context.Context, req *pb.UpdateAlbumRequest) (*pb.Album, error) {
	a := &models.Album{ID: req.Id, Name: req.Name, Description: req.Description, MaxPhotos: int(req.MaxPhotos)}
	if err := s.svc.Update(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toAlbum(a), nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteAlbumRequest) (*emptypb.Empty, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
