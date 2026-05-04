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
	svc AlbumService
}

func NewServer(svc AlbumService) *Server {
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

func toAlbumPhoto(ap *models.AlbumPhoto) *pb.AlbumPhoto {
	return &pb.AlbumPhoto{
		AlbumId:    ap.AlbumID,
		PhotoId:    ap.PhotoID,
		OrderIndex: int32(ap.OrderIndex),
		CreatedAt:  ap.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (s *Server) Create(ctx context.Context, req *pb.CreateAlbumRequest) (*pb.Album, error) {
	a := &models.Album{
		TripID:      req.TripId,
		Name:        req.Name,
		Description: req.Description,
		MaxPhotos:   int(req.MaxPhotos),
	}
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

func (s *Server) GetByTrip(ctx context.Context, req *pb.GetAlbumByTripRequest) (*pb.Album, error) {
	albums, err := s.svc.GetByTrip(ctx, req.TripId)
	if err != nil || len(albums) == 0 {
		return nil, status.Error(codes.NotFound, "album not found")
	}
	return toAlbum(&albums[0]), nil
}

func (s *Server) Update(ctx context.Context, req *pb.UpdateAlbumRequest) (*pb.Album, error) {
	a := &models.Album{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		MaxPhotos:   int(req.MaxPhotos),
	}
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

func (s *Server) AddPhoto(ctx context.Context, req *pb.AddPhotoRequest) (*pb.AlbumPhoto, error) {
	if err := s.svc.AddPhoto(ctx, req.AlbumId, req.PhotoId, int16(req.Order)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AlbumPhoto{
		AlbumId:    req.AlbumId,
		PhotoId:    req.PhotoId,
		OrderIndex: req.Order,
	}, nil
}

func (s *Server) RemovePhoto(ctx context.Context, req *pb.RemovePhotoRequest) (*emptypb.Empty, error) {
	if err := s.svc.RemovePhoto(ctx, req.AlbumId, req.PhotoId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetPhotos(ctx context.Context, req *pb.GetAlbumPhotosRequest) (*pb.GetAlbumPhotosResponse, error) {
	photos, err := s.svc.GetPhotos(ctx, req.AlbumId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pbPhotos []*pb.AlbumPhoto
	for _, p := range photos {
		pbPhotos = append(pbPhotos, toAlbumPhoto(&p))
	}
	return &pb.GetAlbumPhotosResponse{Photos: pbPhotos}, nil
}
