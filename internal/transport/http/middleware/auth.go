package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"

	jwtpkg "gopress/pkg/jwt"
)

type ctxKey string

const (
	ctxUserIDKey   ctxKey = "userId"
	ctxUsernameKey ctxKey = "username"
)

func RequireAuth(jwtManager *jwtpkg.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := jwtManager.ParseToken(cookie.Value)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserIDKey, userID)
		ctx = context.WithValue(ctx, ctxUsernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserIDKey)
	if v == nil {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
