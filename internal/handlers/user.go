package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, creds *models.Creds) (string, error)
	Authenticate(ctx context.Context, creds *models.Creds) (string, error)
}
type UserHandler struct {
	authSvc AuthService
	logger  *zap.Logger
}

func NewUserHandler(authSvc AuthService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var c models.Creds
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if c.Login == "" || c.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := uh.authSvc.Register(r.Context(), &c)
	if err != nil {
		if errors.Is(err, models.ErrUserAlreadyExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		uh.logger.Error("failed to register user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var c models.Creds
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if c.Login == "" || c.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := uh.authSvc.Authenticate(r.Context(), &c)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrInvalidCredentials) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		uh.logger.Error("failed to authenticate user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)
	w.WriteHeader(http.StatusOK)
}

func setAuthCookie(w http.ResponseWriter, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 24),
	})
}
