package album

import (
	"context"
	"guidely-app/pkg/models"
	"guidely-app/pkg/pb/album"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	album.UnimplementedAlbumServiceServer
	svc *Service
}

func NewServer(svc *Service) *Server {
	return &Server{svc: svc}
}

func toProtoAlbum(a *models.Album) *album.Album {
	return &album.Album{
		Id:           a.ID,
		TripId:       a.TripID,
		Name:         a.Name,
		Description:  a.Description,
		CoverPhotoId: a.CoverPhotoID,
		MaxPhotos:    int32(a.MaxPhotos),
	}
}

func (s *Server) Create(ctx context.Context, req *album.CreateAlbumRequest) (*album.Album, error) {
	a := &models.Album{
		TripID:      req.TripId,
		Name:        req.Name,
		Description: req.Description,
		MaxPhotos:   int(req.MaxPhotos),
	}
	if err := s.svc.Create(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoAlbum(a), nil
}

func (s *Server) Get(ctx context.Context, req *album.GetAlbumRequest) (*album.Album, error) {
	a, err := s.svc.GetByID(ctx, req.Id)
	if err != nil || a == nil {
		return nil, status.Error(codes.NotFound, "album not found")
	}
	return toProtoAlbum(a), nil
}

func (s *Server) Update(ctx context.Context, req *album.UpdateAlbumRequest) (*album.Album, error) {
	a := &models.Album{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		MaxPhotos:   int(req.MaxPhotos),
	}
	if err := s.svc.Update(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoAlbum(a), nil
}

func (s *Server) Delete(ctx context.Context, req *album.DeleteAlbumRequest) (*emptypb.Empty, error) {
	if err := s.svc.Delete(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
