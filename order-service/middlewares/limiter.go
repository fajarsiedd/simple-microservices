package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	Rate      rate.Limit
	Burst     int
	ExpiresIn time.Duration
}

func (c *RateLimiterConfig) Init() echo.MiddlewareFunc {
	config := middleware.RateLimiterMemoryStoreConfig{
		Rate:      c.Rate,
		Burst:     c.Burst,
		ExpiresIn: c.ExpiresIn,
	}

	memoryStore := middleware.NewRateLimiterMemoryStoreWithConfig(config)

	return middleware.RateLimiter(memoryStore)
}
