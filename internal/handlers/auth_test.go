package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "guidely-app/pkg/pb/auth"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// mockAuthClient реализует интерфейс pb.AuthServiceClient для тестов
type mockAuthClient struct {
	registerFunc func(*pb.RegisterRequest, ...grpc.CallOption) (*pb.RegisterResponse, error)
	loginFunc    func(*pb.LoginRequest, ...grpc.CallOption) (*pb.LoginResponse, error)
	logoutFunc   func(*pb.LogoutRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	getUserFunc  func(*pb.GetUserRequest, ...grpc.CallOption) (*pb.User, error)
	// остальные методы при необходимости можно добавить как заглушки
}

func (m *mockAuthClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(in, opts...)
	}
	return nil, errors.New("Register not implemented")
}

func (m *mockAuthClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(in, opts...)
	}
	return nil, errors.New("Login not implemented")
}

func (m *mockAuthClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if m.logoutFunc != nil {
		return m.logoutFunc(in, opts...)
	}
	return nil, errors.New("Logout not implemented")
}

func (m *mockAuthClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(in, opts...)
	}
	return nil, errors.New("GetUser not implemented")
}

// Заглушки для оставшихся методов интерфейса (чтобы компилятор не ругался)
func (m *mockAuthClient) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest, opts ...grpc.CallOption) (*pb.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockAuthClient) UploadAvatar(ctx context.Context, opts ...grpc.CallOption) (pb.AuthService_UploadAvatarClient, error) {
	return nil, errors.New("not implemented")
}
func (m *mockAuthClient) GetAvatar(ctx context.Context, in *pb.GetAvatarRequest, opts ...grpc.CallOption) (pb.AuthService_GetAvatarClient, error) {
	return nil, errors.New("not implemented")
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockClient := &mockAuthClient{
		registerFunc: func(req *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
			return &pb.RegisterResponse{UserId: 1, Message: "user created"}, nil
		},
	}
	handler := NewAuthHandler(mockClient)

	reqBody := map[string]string{"login": "test@example.com", "password": "12345678", "nickname": "tester"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, float64(1), resp["user_id"])
	assert.Equal(t, "user created", resp["message"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockClient := &mockAuthClient{
		loginFunc: func(req *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
			return nil, errors.New("invalid credentials")
		},
	}
	handler := NewAuthHandler(mockClient)

	reqBody := map[string]string{"login": "test@example.com", "password": "wrong"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
