# 测试文档

本文档说明项目的测试策略、测试类型和如何运行测试。

## 测试类型

项目包含三种类型的测试：

### 1. 单元测试 (Unit Tests)

单元测试针对独立的函数和方法进行测试，使用mock对象隔离依赖。

**位置**: `tests/unit/`

**特点**:
- 快速执行
- 不依赖外部服务（数据库、Redis等）
- 使用mock对象模拟依赖
- 测试覆盖率目标：80%+

**运行方式**:
```bash
# 运行所有单元测试
go test ./tests/unit/... -v

# 运行特定包的单元测试
go test ./tests/unit/user_service_test.go -v

# 查看测试覆盖率
go test ./tests/unit/... -cover
```

**示例**:
- `user_service_test.go`: 测试用户服务的注册、登录等功能
- 使用 `testify/mock` 创建mock对象
- 使用 `testify/assert` 进行断言

**注意**: 当前 `TestUserService_Register` 和 `TestUserService_Login` 测试被跳过，因为 `UserService` 使用具体类型 `*repository.UserRepository` 而非接口，无法直接使用 mock。这些功能通过集成测试 (`tests/integration/api_test.go`) 进行验证。如需进行单元测试，需要重构 `UserService` 使用接口。

### 2. 集成测试 (Integration Tests)

集成测试测试完整的API端点，需要真实的数据库和Redis连接。

**位置**: `tests/integration/`

**特点**:
- 测试完整的请求-响应流程
- 需要测试数据库和Redis
- 测试真实的数据库操作
- 验证API端点的正确性

**运行方式**:
```bash
# 运行集成测试（需要先启动测试数据库和Redis）
go test ./tests/integration/... -v

# 使用测试数据库
export DB_NAME=enterprise_blog_test
go test ./tests/integration/... -v
```

**环境要求**:
- PostgreSQL测试数据库
- Redis测试实例（可选）
- 测试数据会自动清理

**示例**:
- `api_test.go`: 测试用户注册、登录、文章CRUD等API端点
- 使用 `httptest` 创建HTTP测试请求
- 验证HTTP状态码和响应格式

### 3. E2E测试 (End-to-End Tests)

E2E测试使用Playwright在真实浏览器中测试完整的用户流程。

**位置**: `frontend/tests/e2e/`（测试文件）和 `frontend/playwright.config.ts`（配置文件）

**特点**:
- 在真实浏览器中运行
- 测试完整的用户交互流程
- 验证前端和后端的集成
- 可以测试UI和用户体验

**运行方式**:
```bash
# 首次运行需要先安装前端依赖和Playwright浏览器
make install-frontend

# 运行E2E测试（需要先启动前后端服务）
make test-e2e
# 或
cd frontend && npm run test:e2e

# 使用UI模式运行测试（可视化）
cd frontend && npm run test:e2e:ui

# 查看测试报告
cd frontend && npx playwright show-report
```

**环境要求**:
- 前端开发服务器运行在 `http://localhost:3000`
- 后端API服务器运行在 `http://localhost:8080`
- 数据库和Redis服务已启动
- **重要**: 确保数据库中有测试数据（至少有一篇文章），否则部分测试会被跳过

**测试文件**:
- `frontend/tests/e2e/auth.spec.ts`: 用户认证测试（注册、登录）
- `frontend/tests/e2e/articles.spec.ts`: 文章功能测试（列表、详情、创建、搜索）
- `frontend/tests/e2e/comments.spec.ts`: 评论功能测试（查看、发表）

**配置文件**:
- `frontend/playwright.config.ts`: Playwright 配置文件

**测试覆盖**:
- 用户认证流程（注册、登录、登录失败）
- 文章功能（列表、详情、创建、搜索）
- 评论功能（查看、发表）
- 支持多个浏览器（Chromium、Firefox、WebKit）

## 测试覆盖率

### 目标覆盖率

- **单元测试**: 80%+（当前部分测试被跳过，实际覆盖率较低）
- **集成测试**: 覆盖所有主要API端点
- **E2E测试**: 覆盖主要用户流程

### 查看覆盖率

```bash
# 生成覆盖率报告（所有包）
make test-coverage
# 或
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 查看特定包的覆盖率
go test ./internal/models/... -cover
go test ./internal/services/... -cover

# 查看覆盖率详情
go tool cover -func=coverage.out
```

**注意**: 
- 当前 `tests/unit/` 中的部分测试被跳过（因为需要接口重构），所以单元测试覆盖率显示较低
- 实际业务逻辑通过集成测试进行验证，覆盖率更高
- `TestUserPasswordHashing` 测试正常运行，覆盖了密码哈希功能

## 测试最佳实践

### 1. 单元测试

- 使用mock对象隔离依赖
- 测试边界条件和错误情况
- 保持测试独立，不依赖执行顺序
- 使用表驱动测试（table-driven tests）

### 2. 集成测试

- 使用测试数据库，避免影响生产数据
- 每个测试后清理测试数据
- 测试真实的数据库操作
- 验证完整的请求-响应流程

### 3. E2E测试

- 测试关键用户流程
- 使用数据测试ID（data-testid）定位元素
- 等待异步操作完成
- 截图和视频记录失败的测试

## CI/CD集成

### GitHub Actions示例

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test ./tests/unit/... -v

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:12
        env:
          POSTGRES_DB: enterprise_blog_test
      redis:
        image: redis:6
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test ./tests/integration/... -v

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm install
      - run: npx playwright install
      - run: npm run test:e2e
```

## 监控指标

项目集成了Prometheus监控指标，可以通过 `/metrics` 端点访问。

### 可用指标

- **HTTP请求指标**:
  - `http_requests_total`: HTTP请求总数
  - `http_request_duration_seconds`: HTTP请求持续时间
  - `http_requests_in_flight`: 当前正在处理的请求数

- **数据库指标**:
  - `db_queries_total`: 数据库查询总数
  - `db_query_duration_seconds`: 数据库查询持续时间

- **Redis指标**:
  - `redis_operations_total`: Redis操作总数
  - `redis_operation_duration_seconds`: Redis操作持续时间

- **业务指标**:
  - `user_registrations_total`: 用户注册总数
  - `article_creations_total`: 文章创建总数
  - `comment_creations_total`: 评论创建总数
  - `article_likes_total`: 文章点赞总数
  - `active_users`: 当前活跃用户数

### 访问指标

```bash
# 访问Prometheus metrics端点
curl http://localhost:8080/metrics
```

### 配置Prometheus

在 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'enterprise-blog'
    static_configs:
      - targets: ['localhost:8080']
```

## 故障排查

### 单元测试失败

- 检查mock对象的设置是否正确
- 验证测试数据的准备
- 检查断言条件

### 集成测试失败

- 确认数据库连接正常
- 检查测试数据是否已清理
- 验证API端点是否可访问

### E2E测试失败

- 确认前后端服务已启动
- 检查浏览器是否已安装
- 查看测试截图和视频
- 验证页面元素选择器是否正确

## 贡献指南

添加新功能时，请同时添加相应的测试：

1. **单元测试**: 为新服务方法添加单元测试
2. **集成测试**: 为新API端点添加集成测试
3. **E2E测试**: 为新用户流程添加E2E测试

保持测试代码的质量和可维护性，遵循项目的测试最佳实践。

