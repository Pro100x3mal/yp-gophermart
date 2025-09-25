package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/middleware"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/validate"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"go.uber.org/zap"
)

type OrdersService interface {
	LoadOrder(ctx context.Context, userID int64, num string) error
	ListOrders(ctx context.Context, userID int64) ([]models.Order, error)
}

type OrdersHandler struct {
	ordersSvc OrdersService
	logger    *zap.Logger
}

func NewOrdersHandler(ordersSvc OrdersService, logger *zap.Logger) *OrdersHandler {
	return &OrdersHandler{
		ordersSvc: ordersSvc,
		logger:    logger.With(zap.String("handler", "orders")),
	}
}

func (oh *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	number := strings.TrimSpace(string(body))
	if number == "" || !validate.ValidLuhn(number) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	if err = oh.ordersSvc.LoadOrder(r.Context(), userID, number); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		if errors.Is(err, models.ErrOrderAlreadyUploadedBySameUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, models.ErrOrderBelongsToAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		oh.logger.Error("failed to load order", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (oh *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	orders, err := oh.ordersSvc.ListOrders(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		oh.logger.Error("failed to get orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(orders); err != nil {
		oh.logger.Error("failed to encode orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
