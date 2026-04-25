package auth

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"guidely-app/internal/auth/repository"
	"guidely-app/pkg/models"
	"guidely-app/pkg/pb/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	auth.UnimplementedAuthServiceServer
	svc *Service
}

func NewServer(pool *pgxpool.Pool) auth.AuthServiceServer {
	adapter := &repository.PgxPoolAdapter{Pool: pool}
	userRepo := repository.NewUserRepo(adapter)
	sessRepo := repository.NewSessionRepo(adapter)
	svc := NewService(userRepo, sessRepo)
	return &server{svc: svc}
}

func userToPB(u *models.User) *auth.User {
	return &auth.User{
		Id:         u.ID,
		Login:      u.Login,
		Nickname:   u.Nickname,
		AvatarUrl:  u.AvatarURL,
		Country:    u.Country,
		City:       u.City,
		About:      u.About,
		HasReviews: u.HasReviews,
		CreatedAt:  u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  u.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	user, token, err := s.svc.Register(ctx, RegisterInput{
		Login:    req.Login,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &auth.RegisterResponse{UserId: user.ID, Message: "user created"}, nil
}

func (s *server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, token, err := s.svc.Login(ctx, LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &auth.LoginResponse{
		UserId:    user.ID,
		Token:     token,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
	}, nil
}

func (s *server) Logout(ctx context.Context, req *auth.LogoutRequest) (*emptypb.Empty, error) {
	if err := s.svc.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *server) GetMe(ctx context.Context, req *auth.GetMeRequest) (*auth.User, error) {
	user, err := s.svc.GetUserByID(ctx, req.UserId)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return userToPB(user), nil
}

func (s *server) UpdateProfile(ctx context.Context, req *auth.UpdateProfileRequest) (*auth.User, error) {
	user, err := s.svc.UpdateProfile(ctx, req.UserId,
		req.Nickname, req.AvatarUrl, req.Country, req.City, req.About)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return userToPB(user), nil
}

func (s *server) UploadAvatar(stream auth.AuthService_UploadAvatarServer) error {
	var userID uint64
	var fileData []byte
	var filename string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if userID == 0 {
			userID = chunk.UserId
		}
		fileData = append(fileData, chunk.Chunk...)
		filename = chunk.FileName
	}
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	newFilename := uuid.New().String() + ext
	uploadDir := "./uploads/avatars"
	os.MkdirAll(uploadDir, 0755)
	filePath := filepath.Join(uploadDir, newFilename)
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	avatarURL := "/uploads/avatars/" + newFilename
	user, err := s.svc.UpdateAvatar(context.Background(), userID, avatarURL)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return stream.SendAndClose(userToPB(user))
}

func (s *server) GetAvatar(req *auth.GetAvatarRequest, stream auth.AuthService_GetAvatarServer) error {
	user, err := s.svc.GetUserByID(context.Background(), req.UserId)
	if err != nil || user == nil || user.AvatarURL == "" {
		return status.Error(codes.NotFound, "avatar not found")
	}
	fpath := strings.TrimPrefix(user.AvatarURL, "/uploads/")
	fpath = filepath.Join("./uploads", fpath)
	file, err := os.Open(fpath)
	if err != nil {
		return status.Error(codes.NotFound, "avatar file not found")
	}
	defer file.Close()
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if err := stream.Send(&auth.GetAvatarResponse{Chunk: buf[:n]}); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}
