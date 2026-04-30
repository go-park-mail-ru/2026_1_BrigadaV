package auth

import (
	"context"

	"guidely-app/pkg/pb/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	auth.UnimplementedAuthServiceServer
	svc *Service
}

func NewServer(svc *Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	user, _, err := s.svc.Register(ctx, RegisterInput{Login: req.Login, Password: req.Password, Nickname: req.Nickname})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &auth.RegisterResponse{UserId: user.ID, Message: "user created"}, nil
}

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, token, err := s.svc.Login(ctx, LoginInput{Login: req.Login, Password: req.Password})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &auth.LoginResponse{UserId: user.ID, Token: token, Nickname: user.Nickname, AvatarUrl: user.AvatarURL}, nil
}

func (s *Server) Logout(ctx context.Context, req *auth.LogoutRequest) (*emptypb.Empty, error) {
	if err := s.svc.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetUser(ctx context.Context, req *auth.GetUserRequest) (*auth.User, error) {
	user, err := s.svc.GetUserByID(ctx, req.UserId)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &auth.User{
		Id:        user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		Role:      user.Role,
	}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *auth.UpdateProfileRequest) (*auth.User, error) {
	user, err := s.svc.UpdateProfile(ctx, req.UserId, req.Nickname, req.AvatarUrl, req.Country, req.City, req.About)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auth.User{
		Id:        user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		Role:      user.Role,
	}, nil
}

func (s *Server) UploadAvatar(stream auth.AuthService_UploadAvatarServer) error {
	// реализация загрузки через стрим
	return nil
}

func (s *Server) GetAvatar(req *auth.GetAvatarRequest, stream auth.AuthService_GetAvatarServer) error {
	return nil
}
