package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type TokenValidator interface {
	Validate(token string) (int64, error)
}

type contextKey string

const userIDKey contextKey = "user_id"

func Auth(logger *zap.Logger, tv TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mLog := logger.With(zap.String("middleware", "auth"))

			c, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			userID, err := tv.Validate(c.Value)
			if err != nil || userID <= 0 {
				mLog.Error("invalid jwt token", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, false
	}
	id, ok := val.(int64)
	return id, ok
}
