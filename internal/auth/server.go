package auth

import (
	"context"

	pb "guidely-app/pkg/pb/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	svc AuthService
}

func NewServer(svc AuthService) *Server {
	return &Server{svc: svc}
}
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, token, err := s.svc.Register(ctx, RegisterInput{Login: req.Login, Password: req.Password, Nickname: req.Nickname})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.RegisterResponse{UserId: user.ID, Message: "user created", Token: token}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, token, err := s.svc.Login(ctx, LoginInput{Login: req.Login, Password: req.Password})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.LoginResponse{UserId: user.ID, Token: token, Nickname: user.Nickname, AvatarUrl: user.AvatarURL}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	if err := s.svc.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := s.svc.GetUserByID(ctx, req.UserId)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &pb.User{
		Id:        user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		Role:      user.Role,
	}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.User, error) {
	user, err := s.svc.UpdateProfile(ctx, req.UserId, req.Nickname, req.AvatarUrl, req.Country, req.City, req.About)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.User{
		Id:        user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		Role:      user.Role,
	}, nil
}

func (s *Server) UploadAvatar(stream pb.AuthService_UploadAvatarServer) error { return nil }
func (s *Server) GetAvatar(req *pb.GetAvatarRequest, stream pb.AuthService_GetAvatarServer) error {
	return nil
}
