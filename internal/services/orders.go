package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
)

type OrdersRepository interface {
	InsertOrder(ctx context.Context, userID int64, num string) error
	GetOrderOwnerID(ctx context.Context, num string) (int64, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]models.Order, error)
}

type OrdersService struct {
	repo OrdersRepository
}

func NewOrdersService(repo OrdersRepository) *OrdersService {
	return &OrdersService{
		repo: repo,
	}
}

func (os *OrdersService) LoadOrder(ctx context.Context, userID int64, num string) error {
	if err := os.repo.InsertOrder(ctx, userID, num); err != nil {
		if errors.Is(err, models.ErrOrderExists) {
			ownerID, err := os.repo.GetOrderOwnerID(ctx, num)
			if err != nil {
				return err
			}
			if ownerID == userID {
				return models.ErrOrderAlreadyUploadedBySameUser
			}
			return models.ErrOrderBelongsToAnotherUser
		}
		return fmt.Errorf("failed to load order: %w", err)
	}
	return nil
}

func (os *OrdersService) ListOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	orders, err := os.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user_id: %w", err)
	}
	return orders, nil
}
