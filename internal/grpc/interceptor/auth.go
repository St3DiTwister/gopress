package interceptor

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	jwtpkg "gopress/pkg/jwt"
)

type ctxKey string

const CtxUserIDKey ctxKey = "user_id"
const CtxUsernameKey ctxKey = "username"

type AuthInterceptor struct {
	jwtManager *jwtpkg.Manager

	public map[string]struct{}
}

func NewAuthInterceptor(jwtManager *jwtpkg.Manager, publicMethods []string) *AuthInterceptor {
	m := make(map[string]struct{}, len(publicMethods))
	for _, v := range publicMethods {
		m[v] = struct{}{}
	}
	return &AuthInterceptor{
		jwtManager: jwtManager,
		public:     m,
	}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := a.public[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		token, err := extractBearer(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := a.jwtManager.ParseToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token user_id")
		}

		ctx = context.WithValue(ctx, CtxUserIDKey, userID)
		ctx = context.WithValue(ctx, CtxUsernameKey, claims.Username)

		return handler(ctx, req)
	}
}

func extractBearer(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}

	v := strings.TrimSpace(values[0])
	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "empty token")
	}
	return token, nil
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(CtxUserIDKey)
	id, ok := v.(uuid.UUID)
	return id, ok
}
