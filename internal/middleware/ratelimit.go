package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s:%s", clientIP, c.Request.URL.Path)

		ctx := context.Background()
		
		// 获取当前计数
		count, err := database.RedisClient.Get(ctx, key).Int()
		if err != nil && err.Error() != "redis: nil" {
			c.Next()
			return
		}

		if count >= limit {
			c.JSON(http.StatusTooManyRequests, models.Error(429, "too many requests"))
			c.Abort()
			return
		}

		// 增加计数
		pipe := database.RedisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(limit-count-1))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

		c.Next()
	}
}

