package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopress/api/proto/auth"
	"gopress/internal/models"
	"gopress/internal/repository"
	jwtpkg "gopress/pkg/jwt"
	"gopress/pkg/password"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	userRepo   repository.UserRepo
	jwtManager *jwtpkg.Manager
}

func NewAuthServer(userRepo repository.UserRepo, jwtManager *jwtpkg.Manager) *AuthServer {
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

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: hashed,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &auth.RegisterResponse{
		Id:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	if !password.Check(user.Password, req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &auth.LoginResponse{
		Token: token,
	}, nil
}
