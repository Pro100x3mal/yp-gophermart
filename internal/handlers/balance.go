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

type BalanceService interface {
	CalculateBalance(ctx context.Context, userID int64) (*models.Balance, error)
	WithdrawFunds(ctx context.Context, userID int64, wd *models.WithdrawReq) error
	ListWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

type BalanceHandler struct {
	balanceSvc BalanceService
	logger     *zap.Logger
}

func NewBalanceHandler(balanceSvc BalanceService, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		balanceSvc: balanceSvc,
		logger:     logger.With(zap.String("handler", "balance")),
	}
}

func (bh *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	balance, err := bh.balanceSvc.CalculateBalance(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		bh.logger.Error("failed to get balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(balance); err != nil {
		bh.logger.Error("failed to encode balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (bh *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var wd models.WithdrawReq
	if err = json.Unmarshal(body, &wd); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !validate.ValidLuhn(wd.Order) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	if wd.Sum <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = bh.balanceSvc.WithdrawFunds(r.Context(), userID, &wd); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		if errors.Is(err, models.ErrPaymentRequired) {
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
			return
		}
		if errors.Is(err, models.ErrWithdrawalOrderExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		bh.logger.Error("failed to withdraw funds", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (bh *BalanceHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	withdrawals, err := bh.balanceSvc.ListWithdrawals(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			return
		}
		bh.logger.Error("failed to list withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(withdrawals); err != nil {
		bh.logger.Error("failed to encode withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
