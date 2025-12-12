package services

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopress/api/proto/auth"
	"gopress/internal/app/ports"
	"gopress/internal/domain/user"
	jwtpkg "gopress/pkg/jwt"
	"gopress/pkg/password"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	userRepo   ports.UserRepo
	jwtManager *jwtpkg.Manager
}

func NewAuthServer(userRepo ports.UserRepo, jwtManager *jwtpkg.Manager) *AuthServer {
	return &AuthServer{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "empty fields")
	}

	hashed, err := password.Hash(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	u := &user.User{
		Email:    req.Email,
		Username: req.Username,
		Password: hashed,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, status.Error(codes.Internal, "failed to create u")
	}

	return &auth.RegisterResponse{
		Id:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	u, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	if u == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	if !password.Check(u.Password, req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	token, err := s.jwtManager.GenerateToken(u.ID, u.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &auth.LoginResponse{
		Token: token,
	}, nil
}
