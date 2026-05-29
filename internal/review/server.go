package review

import (
	"context"
	"strings"
	"time"

	"guidely-app/pkg/models"
	pb "guidely-app/pkg/pb/review"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedReviewServiceServer
	svc ReviewService
}

func NewServer(svc ReviewService) *Server {
	return &Server{svc: svc}
}

func toReviewResponse(r *models.Review) *pb.ReviewResponse {
	resp := &pb.ReviewResponse{
		Id:        r.ID,
		UserId:    r.UserID,
		PlaceId:   r.PlaceID,
		Rating:    int32(r.Rating),
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
	if r.Title != nil {
		resp.Title = r.Title
	}
	if r.VisitDate != nil {
		visitDate := r.VisitDate.Format("2006-01-02")
		resp.VisitDate = &visitDate
	}
	return resp
}

func toReviewWithAuthorMessages(reviews []models.ReviewWithAuthor) []*pb.ReviewWithAuthor {
	var result []*pb.ReviewWithAuthor
	for _, r := range reviews {
		msg := &pb.ReviewWithAuthor{
			Id:        r.ID,
			Rating:    int32(r.Rating),
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			Author: &pb.Author{
				Id:       r.Author.ID,
				Nickname: r.Author.Nickname,
				Avatar:   r.Author.Avatar,
			},
		}
		if r.Title != nil {
			msg.Title = r.Title
		}
		result = append(result, msg)
	}
	return result
}

// isDuplicateError определяет, является ли ошибка нарушением уникальности.
// Таблица review имеет UNIQUE(user_id, place_id).
func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "unique_violation") ||
		strings.Contains(msg, "duplicate key")
}

func (s *Server) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.ReviewResponse, error) {
	var visitDate *time.Time
	if req.VisitDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.VisitDate)
		if err == nil {
			visitDate = &parsed
		}
	}
	input := CreateReviewInput{
		UserID:    req.UserId,
		PlaceID:   req.PlaceId,
		Title:     req.Title,
		Rating:    int16(req.Rating),
		Comment:   req.Comment,
		VisitDate: visitDate,
	}
	review, err := s.svc.Create(ctx, input)
	if err != nil {
		// Проверяем на дублирующийся отзыв
		if isDuplicateError(err) {
			return nil, status.Error(codes.AlreadyExists, "you have already reviewed this place")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toReviewResponse(review), nil
}

func (s *Server) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*emptypb.Empty, error) {
	if err := s.svc.Delete(ctx, req.UserId, req.ReviewId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetReviewsByPlace(ctx context.Context, req *pb.GetReviewsByPlaceRequest) (*pb.GetReviewsByPlaceResponse, error) {
	reviews, err := s.svc.GetByPlaceIDWithAuthor(ctx, req.PlaceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetReviewsByPlaceResponse{
		Reviews: toReviewWithAuthorMessages(reviews),
	}, nil
}

func (s *Server) CheckUserReview(ctx context.Context, req *pb.CheckUserReviewRequest) (*pb.CheckUserReviewResponse, error) {
	exists, err := s.svc.CheckUserReview(ctx, req.UserId, req.PlaceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CheckUserReviewResponse{Exists: exists}, nil
}
