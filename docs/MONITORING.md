# 监控指标文档

本文档说明项目的Prometheus监控指标配置和使用方法。

## 概述

项目集成了Prometheus监控指标，通过 `/metrics` 端点暴露指标数据，可以用于监控系统性能、业务指标和健康状态。

## 指标端点

### 访问方式

```bash
curl http://localhost:8080/metrics
```

### 响应格式

指标以Prometheus标准格式返回，例如：

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/articles",status="200"} 1234

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/v1/articles",le="0.005"} 1000
http_request_duration_seconds_bucket{method="GET",path="/api/v1/articles",le="0.01"} 1200
...
```

## 可用指标

### HTTP请求指标

#### `http_requests_total`
- **类型**: Counter
- **描述**: HTTP请求总数
- **标签**:
  - `method`: HTTP方法（GET, POST, PUT, DELETE等）
  - `path`: 请求路径
  - `status`: HTTP状态码

#### `http_request_duration_seconds`
- **类型**: Histogram
- **描述**: HTTP请求持续时间（秒）
- **标签**:
  - `method`: HTTP方法
  - `path`: 请求路径
- **分桶**: 默认Prometheus分桶（0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10秒）

#### `http_requests_in_flight`
- **类型**: Gauge
- **描述**: 当前正在处理的HTTP请求数

### 数据库指标

#### `db_queries_total`
- **类型**: Counter
- **描述**: 数据库查询总数
- **标签**:
  - `operation`: 操作类型（SELECT, INSERT, UPDATE, DELETE）
  - `table`: 表名

#### `db_query_duration_seconds`
- **类型**: Histogram
- **描述**: 数据库查询持续时间（秒）
- **标签**:
  - `operation`: 操作类型
  - `table`: 表名
- **分桶**: 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10秒

### Redis指标

#### `redis_operations_total`
- **类型**: Counter
- **描述**: Redis操作总数
- **标签**:
  - `operation`: 操作类型（GET, SET, DEL, INCR等）

#### `redis_operation_duration_seconds`
- **类型**: Histogram
- **描述**: Redis操作持续时间（秒）
- **标签**:
  - `operation`: 操作类型
- **分桶**: 0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1秒

### 业务指标

#### `user_registrations_total`
- **类型**: Counter
- **描述**: 用户注册总数

#### `article_creations_total`
- **类型**: Counter
- **描述**: 文章创建总数

#### `comment_creations_total`
- **类型**: Counter
- **描述**: 评论创建总数

#### `article_likes_total`
- **类型**: Counter
- **描述**: 文章点赞总数

#### `active_users`
- **类型**: Gauge
- **描述**: 当前活跃用户数

## 配置Prometheus

### 1. 安装Prometheus

```bash
# 下载Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64
```

### 2. 配置prometheus.yml

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'enterprise-blog'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### 3. 启动Prometheus

```bash
./prometheus --config.file=prometheus.yml
```

### 4. 访问Prometheus UI

打开浏览器访问 `http://localhost:9090`

## 配置Grafana

### 1. 安装Grafana

```bash
# Ubuntu/Debian
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
sudo apt-get update
sudo apt-get install grafana

# 启动Grafana
sudo systemctl start grafana-server
```

### 2. 添加Prometheus数据源

1. 登录Grafana（默认用户名/密码：admin/admin）
2. 进入 Configuration > Data Sources
3. 添加Prometheus数据源
4. URL设置为 `http://localhost:9090`

### 3. 导入仪表板

可以使用以下PromQL查询创建仪表板：

**HTTP请求速率**:
```promql
rate(http_requests_total[5m])
```

**HTTP请求延迟（P95）**:
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**数据库查询速率**:
```promql
rate(db_queries_total[5m])
```

**活跃用户数**:
```promql
active_users
```

**用户注册速率**:
```promql
rate(user_registrations_total[5m])
```

## 告警规则

### 示例告警规则（alert.rules）

```yaml
groups:
  - name: enterprise_blog_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/sec"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "P95 latency is {{ $value }} seconds"

      - alert: DatabaseSlowQueries
        expr: histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow database queries detected"
          description: "P95 query duration is {{ $value }} seconds"
```

## 使用示例

### 在代码中记录指标

```go
import "enterprise-blog/pkg/metrics"

// 记录HTTP请求
metrics.RecordHTTPRequest("GET", "/api/v1/articles", 200, duration)

// 记录数据库查询
metrics.RecordDBQuery("SELECT", "articles", duration)

// 记录Redis操作
metrics.RecordRedisOperation("GET", duration)

// 记录业务指标
metrics.RecordUserRegistration()
metrics.RecordArticleCreation()
metrics.RecordArticleLike()
```

## 最佳实践

1. **指标命名**: 遵循Prometheus命名规范（使用下划线，单位明确）
2. **标签选择**: 不要使用高基数的标签（如用户ID）
3. **指标类型**: 正确选择Counter、Gauge、Histogram类型
4. **性能影响**: 指标收集应该对性能影响最小
5. **指标清理**: 定期清理不再使用的指标

## 故障排查

### 指标端点无响应

1. 检查服务是否运行
2. 检查 `/metrics` 路由是否注册
3. 检查防火墙设置

### 指标数据不准确

1. 检查指标记录代码是否正确调用
2. 检查标签值是否正确
3. 检查Prometheus抓取配置

### 性能问题

1. 减少指标数量
2. 减少标签数量
3. 使用采样

## 参考资源

- [Prometheus官方文档](https://prometheus.io/docs/)
- [PromQL查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana文档](https://grafana.com/docs/)

