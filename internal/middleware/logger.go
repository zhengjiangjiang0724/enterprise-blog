package middleware

import (
	"time"

	"enterprise-blog/pkg/logger"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		l := logger.GetLogger()
		event := l.Info()
		if len(errorMessage) > 0 {
			event = l.Error()
		}

		if raw != "" {
			path = path + "?" + raw
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Str("ip", clientIP).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP Request")

		// 5xx 错误额外记一条错误日志
		if status >= 500 {
			l2 := logger.GetLogger()
			l2.Error().
				Str("method", method).
				Str("path", path).
				Int("status", status).
				Str("error", errorMessage).
				Msg("HTTP Error")
		}
	}
}

