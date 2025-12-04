package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/models"
	"github.com/go-resty/resty/v2"
)

type AccrualClient struct {
	client *resty.Client
}

func NewAccrualClient(addr string) *AccrualClient {
	return &AccrualClient{
		client: resty.New().
			SetBaseURL(addr).
			SetTimeout(5 * time.Second),
	}
}

func (c *AccrualClient) GetOrderAccrual(ctx context.Context, number string) (*models.AccrualResp, time.Duration, error) {
	var accrualOrder models.AccrualResp
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&accrualOrder).
		Get(fmt.Sprintf("/api/orders/%s", number))
	if err != nil {
		return nil, 0, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &accrualOrder, 0, nil
	case http.StatusNoContent:
		return nil, 0, models.ErrAccrualOrderNotRegistered
	case http.StatusTooManyRequests:
		retry := parseRetryAfter(resp.Header().Get("Retry-After"))
		return nil, retry, models.ErrAccrualOrderTooMany
	default:
		return nil, 0, fmt.Errorf("accrual unexpected status code: %d %s", resp.StatusCode(), resp.Status())

	}
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	if secs, err := time.ParseDuration(h + "s"); err == nil && secs >= 0 {
		return secs
	}
	if t, err := httpTime(h); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

func httpTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid http time: %q", s)
}
