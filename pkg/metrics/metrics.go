// Package metrics 提供 Prometheus 监控指标
package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP请求总数（按方法、路径、状态码）
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求持续时间（按方法、路径）
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 活跃HTTP请求数
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// 数据库查询总数
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	// 数据库查询持续时间
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	// Redis操作总数
	redisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation"},
	)

	// Redis操作持续时间
	redisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
		},
		[]string{"operation"},
	)

	// 业务指标：用户注册数
	userRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	// 业务指标：文章创建数
	articleCreationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "article_creations_total",
			Help: "Total number of articles created",
		},
	)

	// 业务指标：评论创建数
	commentCreationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "comment_creations_total",
			Help: "Total number of comments created",
		},
	)

	// 业务指标：文章点赞数
	articleLikesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "article_likes_total",
			Help: "Total number of article likes",
		},
	)

	// 业务指标：当前在线用户数（示例）
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of currently active users",
		},
	)
)

// RecordHTTPRequest 记录HTTP请求指标
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	status := prometheus.Labels{"method": method, "path": path, "status": string(rune(statusCode))}
	httpRequestsTotal.With(status).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordHTTPRequestStart 开始记录HTTP请求（用于计算in-flight）
func RecordHTTPRequestStart() {
	httpRequestsInFlight.Inc()
}

// RecordHTTPRequestEnd 结束记录HTTP请求
func RecordHTTPRequestEnd() {
	httpRequestsInFlight.Dec()
}

// RecordDBQuery 记录数据库查询指标
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordRedisOperation 记录Redis操作指标
func RecordRedisOperation(operation string, duration time.Duration) {
	redisOperationsTotal.WithLabelValues(operation).Inc()
	redisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordUserRegistration 记录用户注册
func RecordUserRegistration() {
	userRegistrationsTotal.Inc()
}

// RecordArticleCreation 记录文章创建
func RecordArticleCreation() {
	articleCreationsTotal.Inc()
}

// RecordCommentCreation 记录评论创建
func RecordCommentCreation() {
	commentCreationsTotal.Inc()
}

// RecordArticleLike 记录文章点赞
func RecordArticleLike() {
	articleLikesTotal.Inc()
}

// SetActiveUsers 设置当前活跃用户数
func SetActiveUsers(count float64) {
	activeUsers.Set(count)
}

// MetricsMiddleware Gin中间件，自动记录HTTP请求指标
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		RecordHTTPRequestStart()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		RecordHTTPRequest(c.Request.Method, c.FullPath(), statusCode, duration)
		RecordHTTPRequestEnd()
	}
}

