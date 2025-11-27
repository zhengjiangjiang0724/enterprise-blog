# 企业级博客系统

一个基于Go语言开发的高性能、可扩展的企业级博客系统。

## 功能特性

- ✅ 用户认证与授权（JWT）
- ✅ 文章管理（CRUD）
- ✅ 分类和标签系统
- ✅ 评论功能
- ✅ 全文搜索
- ✅ 文件上传
- ✅ Redis缓存
- ✅ 日志记录
- ✅ 限流保护
- ✅ 数据库迁移

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL
- **缓存**: Redis
- **认证**: JWT
- **日志**: Zerolog

## 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6+

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

项目提供了 docker-compose 方案，一次性启动 PostgreSQL、Redis、后端 API 和前端：

```bash
docker-compose up -d postgres redis app frontend
```

- 后端 API: `http://localhost:8080/api/v1`
- 前端 Web: `http://localhost:3000`

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

## 性能测试

运行性能测试：

```bash
go test -bench=. -benchmem ./tests/...
# 或
make benchmark
```

详细的性能测试报告请参考：[性能测试报告](./tests/performance_report.md)

## 文档

- [架构设计文档](./docs/architecture.md) - 详细的架构设计说明
- [技术总结](./docs/technical_summary.md) - 技术实现和挑战解决方案
- [API文档](./docs/API.md) - 完整的API接口文档
- [快速开始](./docs/QUICKSTART.md) - 快速上手指南
- [挑战与解决方案](./docs/CHALLENGES.md) - 开发中的挑战和解决方案
- [项目总览](./docs/PROJECT_OVERVIEW.md) - 项目整体介绍
- [前端架构文档](./docs/frontend_architecture.md) - 前端项目结构与技术说明
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
- XSS防护
- API限流

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

