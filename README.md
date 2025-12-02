# 企业级博客系统

一个基于Go语言开发的高性能、可扩展的企业级博客系统。

## 功能特性

- ✅ 用户认证与授权（JWT，基于角色的访问控制）
  - 邮箱密码登录
  - 手机号验证码登录（自动创建用户）
- ✅ 文章管理（CRUD）+ 文章状态管理（草稿 / 待审核 / 已发布 / 已归档）
- ✅ 评论功能（游客 / 登录用户评论，分页展示）
- ✅ 点赞 / 本地收藏、阅读量统计（Redis 缓存 + 定时回刷）
- ✅ **Elasticsearch 全文搜索**（完全使用Elasticsearch，支持多字段搜索、筛选、排序）
- ✅ **图片上传和管理功能**（支持JPEG、PNG、GIF、WebP格式，图片列表、搜索、标签管理）
- ✅ 分类和标签系统（自动生成 ID，文章可按分类/标签筛选）
- ✅ Redis 缓存（文章详情 & 列表缓存、计数缓冲）
- ✅ 日志记录与访问日志（Zerolog）
- ✅ 限流保护（基于中间件的 API 级限流）
- ✅ 数据库迁移与连接池配置（可通过环境变量调优）
- ✅ React 前端：文章列表 / 详情、Markdown 编辑与预览、草稿箱 & 审核 / 发布流程、评论区、点赞/收藏
- ✅ 后台管理系统（仅管理员）：
  - 用户管理：查看用户列表、查看详情、修改角色与状态（启用/禁用）
  - 文章管理：后台文章列表、详情、状态切换（草稿 / 待审核 / 已发布 / 已归档）、删除
  - 分类与标签管理：在后台创建 / 更新 / 删除分类与标签
  - 仪表盘：用户数、文章数、阅读/点赞总数、评论数、今日发布数等核心统计
  - 系统配置查看：服务器 / 数据库 / Redis / JWT / 日志 / 上传配置（当前为只读视图）

## 技术栈

- **后端语言**: Go 1.21+
- **Web 框架**: Gin
- **数据库**: PostgreSQL
- **搜索引擎**: Elasticsearch（全文搜索）
- **缓存**: Redis（数据缓存 + 计数缓冲）
- **认证**: JWT
- **日志**: Zerolog
- **ORM**: GORM（结合原生 SQL）
- **前端**: React 18 + TypeScript + React Router + Vite（支持 Markdown 编辑与预览）

## 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6+
- Elasticsearch 8+ (可选，用于全文搜索)

### 安装依赖

```bash
go mod download
```

### 配置数据库

创建PostgreSQL数据库：

```sql
CREATE DATABASE enterprise_blog;
```

### 配置环境变量

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

### 运行数据库迁移

```bash
go run cmd/migrate/main.go
```

### 启动服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

### 使用 Docker Compose 启动前后端

项目提供了 docker-compose 方案，一次性启动 PostgreSQL、Redis、Elasticsearch、后端 API 和前端：

```bash
docker-compose up -d
```

或者只启动部分服务：

```bash
docker-compose up -d postgres redis elasticsearch app frontend
```

**服务说明**:
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Elasticsearch: `localhost:9200`
- 后端 API: `http://localhost:8080/api/v1`
- 前端 Web: `http://localhost:3000`

**配置说明**:
- Elasticsearch 配置通过环境变量 `ELASTICSEARCH_URL` 和 `ELASTICSEARCH_ENABLED` 控制
- 图片上传目录通过环境变量 `UPLOAD_DIR` 配置（默认：`./uploads/images`）
- 所有配置都可以通过环境变量或 `.env` 文件设置

## 使用Makefile

项目提供了Makefile来简化常用操作：

```bash
make build     # 构建应用
make run       # 运行服务器
make test      # 运行测试
make migrate   # 运行数据库迁移
make benchmark # 运行性能测试
make clean     # 清理构建产物
```

## API文档

详细的API文档请参考：[API文档](./docs/API.md)

## 测试

### 单元测试

运行单元测试：

```bash
make test-unit
# 或
go test -v ./tests/unit/... -cover
```

### 集成测试

运行集成测试（需要测试数据库）：

```bash
make test-integration
# 或
go test -v ./tests/integration/...
```

### E2E测试

运行端到端测试（需要启动前后端服务）：

```bash
make test-e2e
# 或
cd frontend && npm run test:e2e
```

### 测试覆盖率

生成测试覆盖率报告：

```bash
make test-coverage
```

详细的测试文档请参考：[测试文档](./docs/TESTING.md)

## 性能测试

运行性能测试：

```bash
go test -bench=. -benchmem ./tests/...
# 或
make benchmark
```

详细的性能测试报告请参考：[性能测试报告](./tests/performance_report.md)

## 监控

项目集成了Prometheus监控指标，可以通过 `/metrics` 端点访问。

### 访问指标

```bash
curl http://localhost:8080/metrics
```

### 配置Prometheus

详细的监控配置请参考：[监控文档](./docs/MONITORING.md)

## 文档

- [架构设计文档](./docs/architecture.md) - 详细的架构设计说明
- [技术总结](./docs/technical_summary.md) - 技术实现和挑战解决方案
- [API文档](./docs/API.md) - 完整的API接口文档
- [快速开始](./docs/QUICKSTART.md) - 快速上手指南
- [挑战与解决方案](./docs/CHALLENGES.md) - 开发中的挑战和解决方案
- [项目总览](./docs/PROJECT_OVERVIEW.md) - 项目整体介绍
- [前端架构文档](./docs/frontend_architecture.md) - 前端项目结构与技术说明
- [测试文档](./docs/TESTING.md) - 测试策略和运行指南
- [监控文档](./docs/MONITORING.md) - Prometheus监控指标说明
- [交付清单](./docs/SUMMARY.md) - 完整的交付内容清单

## 核心特性

### 性能
- 单机支持 5000+ QPS
- 平均响应时间 < 5ms
- Redis缓存优化
- 数据库查询优化

### 安全性
- JWT Token认证
- 密码bcrypt加密
- SQL注入防护
  - 参数化查询（所有用户输入通过参数绑定）
  - 白名单验证（ORDER BY 字段名和排序方向）
  - 字符验证和转义（全文搜索查询）
  - 避免字符串拼接 SQL
- XSS防护
- API限流
- 多种登录方式
  - 邮箱密码登录
  - 手机号验证码登录（自动创建用户）
- 密码重置功能（需要验证旧密码）

### 可扩展性
- 分层架构设计
- 支持水平扩展
- 模块化设计
- 易于添加新功能

## 技术亮点

1. **分层架构**: Handler → Service → Repository，职责清晰
2. **JWT认证**: 无状态认证，支持跨域
3. **Redis缓存**: 提升性能，降低数据库压力
4. **软删除**: 数据安全，支持恢复
5. **数据库迁移**: 版本化数据库schema管理
6. **统一错误处理**: 规范化的错误响应格式
7. **中间件**: 认证、日志、限流、CORS

## 项目结构

```
enterprise-blog/
├── cmd/                 # 应用入口
│   ├── server/         # HTTP服务器
│   └── migrate/        # 数据库迁移工具
├── internal/           # 内部代码
│   ├── config/         # 配置管理
│   ├── models/         # 数据模型
│   ├── handlers/       # HTTP处理器
│   ├── services/       # 业务逻辑
│   ├── repository/     # 数据访问层
│   └── middleware/     # 中间件
├── pkg/                # 公共包
│   ├── jwt/           # JWT工具
│   ├── logger/        # 日志工具
│   └── validator/     # 验证器
├── migrations/         # 数据库迁移文件
├── tests/             # 测试文件
└── docs/              # 文档
```

## License

MIT

