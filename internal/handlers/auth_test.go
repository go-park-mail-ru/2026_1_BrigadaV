package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/pkg/config"
	pb "guidely-app/pkg/pb/auth"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockAuthClient struct {
	registerFunc func(*pb.RegisterRequest, ...grpc.CallOption) (*pb.RegisterResponse, error)
	loginFunc    func(*pb.LoginRequest, ...grpc.CallOption) (*pb.LoginResponse, error)
	logoutFunc   func(*pb.LogoutRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	getUserFunc  func(*pb.GetUserRequest, ...grpc.CallOption) (*pb.User, error)
}

func (m *mockAuthClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "Register not implemented")
}
func (m *mockAuthClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "Login not implemented")
}
func (m *mockAuthClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if m.logoutFunc != nil {
		return m.logoutFunc(in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "Logout not implemented")
}
func (m *mockAuthClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "GetUser not implemented")
}
func (m *mockAuthClient) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest, opts ...grpc.CallOption) (*pb.User, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
func (m *mockAuthClient) UploadAvatar(ctx context.Context, opts ...grpc.CallOption) (pb.AuthService_UploadAvatarClient, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
func (m *mockAuthClient) GetAvatar(ctx context.Context, in *pb.GetAvatarRequest, opts ...grpc.CallOption) (pb.AuthService_GetAvatarClient, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockClient := &mockAuthClient{
		registerFunc: func(req *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
			return &pb.RegisterResponse{UserId: 1, Message: "user created"}, nil
		},
		loginFunc: func(req *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
			return &pb.LoginResponse{UserId: 1, Token: "token", Nickname: "tester", AvatarUrl: ""}, nil
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)

	reqBody := `{"login":"test@example.com","password":"12345678","nickname":"tester"}`
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, float64(1), resp["user_id"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockClient := &mockAuthClient{
		loginFunc: func(req *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)

	reqBody := `{"login":"wrong","password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockClient := &mockAuthClient{}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader([]byte(`{invalid`)))
	w := httptest.NewRecorder()
	handler.Register(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_GRPCError(t *testing.T) {
	mockClient := &mockAuthClient{
		registerFunc: func(req *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
			return nil, status.Error(codes.Internal, "db error")
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	reqBody := map[string]string{"login": "test@example.com", "password": "12345678", "nickname": "tester"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.Register(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockClient := &mockAuthClient{}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(`{invalid`)))
	w := httptest.NewRecorder()
	handler.Login(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_GRPCError(t *testing.T) {
	mockClient := &mockAuthClient{
		loginFunc: func(req *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
			return nil, status.Error(codes.Internal, "db error")
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	reqBody := map[string]string{"login": "test@example.com", "password": "12345678"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.Login(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockClient := &mockAuthClient{
		logoutFunc: func(req *pb.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return &emptypb.Empty{}, nil
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token123"})
	w := httptest.NewRecorder()
	handler.Logout(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	mockClient := &mockAuthClient{}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("POST", "/api/logout", nil)
	w := httptest.NewRecorder()
	handler.Logout(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_GRPCError(t *testing.T) {
	mockClient := &mockAuthClient{
		logoutFunc: func(req *pb.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, errors.New("db error")
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token123"})
	w := httptest.NewRecorder()
	handler.Logout(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	mockClient := &mockAuthClient{
		getUserFunc: func(req *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.User, error) {
			return &pb.User{Id: 1, Login: "test@example.com", Nickname: "tester"}, nil
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("GET", "/api/user/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	w := httptest.NewRecorder()
	handler.Me(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var user pb.User
	json.NewDecoder(w.Body).Decode(&user)
	assert.Equal(t, uint64(1), user.Id)
	assert.Equal(t, "test@example.com", user.Login)
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	mockClient := &mockAuthClient{}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("GET", "/api/user/me", nil)
	w := httptest.NewRecorder()
	handler.Me(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Me_UserNotFound(t *testing.T) {
	mockClient := &mockAuthClient{
		getUserFunc: func(req *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.User, error) {
			return nil, status.Error(codes.NotFound, "user not found")
		},
	}
	cfg := &config.Config{SecureCookies: false}
	handler := NewAuthHandler(mockClient, cfg)
	req := httptest.NewRequest("GET", "/api/user/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", uint64(1)))
	w := httptest.NewRecorder()
	handler.Me(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
