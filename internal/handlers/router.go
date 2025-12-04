package handlers

import (
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/jwtmanager"
	"github.com/Pro100x3mal/yp-gophermart.git/internal/infrastructure/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, validator *jwtmanager.JWTManager, ah *AuthHandler, oh *OrdersHandler, bh *BalanceHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Compress(logger))

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", ah.Register)
		r.Post("/login", ah.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(logger, validator))
			r.Post("/orders", oh.CreateOrder)
			r.Get("/orders", oh.GetOrders)
			r.Get("/balance", bh.GetBalance)
			r.Post("/balance/withdraw", bh.Withdraw)
			r.Get("/withdrawals", bh.ListWithdrawals)
		})
	})

	return r
}
