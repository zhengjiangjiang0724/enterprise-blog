# 快速开始指南

## 前置要求

- Go 1.21 或更高版本
- PostgreSQL 12 或更高版本
- Redis 6 或更高版本（可选，但推荐）

## 安装步骤

### 1. 克隆项目

```bash
git clone <repository-url>
cd enterprise-blog
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

创建PostgreSQL数据库：

```sql
CREATE DATABASE enterprise_blog;
```

### 4. 配置环境变量

复制环境变量示例文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，修改数据库和Redis配置：

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=enterprise_blog

REDIS_HOST=localhost
REDIS_PORT=6379
```

### 5. 运行数据库迁移

```bash
go run cmd/migrate/main.go
```

或者使用Makefile：

```bash
make migrate
```

### 6. 启动服务器

```bash
go run cmd/server/main.go
```

或者使用Makefile：

```bash
make run
```

服务器将在 `http://localhost:8080` 启动。

### 7. 测试API

#### 注册用户

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### 登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### 获取文章列表

```bash
curl http://localhost:8080/api/v1/articles
```

## 使用Docker（可选）

### 使用Docker Compose

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: enterprise_blog
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis

volumes:
  postgres_data:
```

启动服务：

```bash
docker-compose up -d
```

## 开发建议

1. **代码格式化**: 使用 `gofmt` 或 `goimports`
2. **代码检查**: 使用 `golint` 或 `golangci-lint`
3. **运行测试**: `make test`
4. **性能测试**: `make benchmark`

## 常见问题

### 数据库连接失败

- 检查PostgreSQL是否运行
- 验证 `.env` 文件中的数据库配置
- 确认数据库已创建

### Redis连接失败

- Redis是可选的，如果未运行，应用会继续运行但不使用缓存
- 检查Redis是否运行：`redis-cli ping`

### 迁移失败

- 确保数据库已创建
- 检查数据库用户权限
- 查看日志了解详细错误信息

## 下一步

- 阅读 [架构设计文档](./architecture.md)
- 查看 [API文档](./API.md)
- 阅读 [技术总结](./technical_summary.md)

