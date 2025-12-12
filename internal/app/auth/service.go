package auth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gopress/internal/app/ports"
	"gopress/internal/domain/user"
	"gopress/pkg/jwt"
	"gopress/pkg/password"
)

var (
	ErrInvalidData   = errors.New("invalid data")
	ErrHashPassword  = errors.New("cannot hash password")
	ErrCreateUser    = errors.New("cannot create user")
	ErrUserNotFound  = errors.New("user not found")
	ErrInternalError = errors.New("internal error")
)

type Service struct {
	repo       ports.UserRepo
	jwtManager *jwt.Manager
}

func NewService(repo ports.UserRepo, jwtManager *jwt.Manager) *Service {
	return &Service{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *Service) Login(ctx context.Context, username, userPassword string) (string, error) {
	if username == "" || userPassword == "" {
		return "", ErrInvalidData
	}
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", ErrInternalError
	}
	if u == nil {
		return "", ErrInvalidData
	}

	if !password.Check(u.Password, userPassword) {
		return "", ErrInvalidData
	}

	token, err := s.jwtManager.GenerateToken(u.ID, u.Username)
	if err != nil {
		return "", ErrInternalError
	}

	return token, nil
}

func (s *Service) Register(ctx context.Context, username, email, userPassword string) (*user.User, error) {
	if username == "" || email == "" || userPassword == "" {
		return nil, ErrInvalidData
	}

	hashed, err := password.Hash(userPassword)
	if err != nil {
		return nil, ErrHashPassword
	}

	u := &user.User{
		Email:    email,
		Username: username,
		Password: hashed,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, ErrCreateUser
	}

	return u, nil
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrInternalError
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	return u, nil
}
