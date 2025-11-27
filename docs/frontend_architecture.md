# 前端项目架构设计文档

## 1. 概述

前端项目是企业级博客系统的独立 Web 客户端，主要职责：

- 提供文章浏览、搜索、详情查看等功能
- 提供用户登录入口（基于后端 JWT）
- 作为后端 REST API 的消费者

技术栈：

- **语言**: TypeScript
- **框架**: React 18 + React Router 6
- **构建工具**: Vite
- **HTTP 客户端**: Axios
- **容器化**: Docker + docker-compose

## 2. 目录结构

```text
frontend/
├── Dockerfile
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── src/
    ├── App.tsx                # 路由与页面布局
    ├── main.tsx               # 入口文件
    ├── styles.css             # 全局样式
    ├── api/
    │   ├── client.ts          # Axios 实例和拦截器
    │   └── types.ts           # 与后端对齐的基础类型
    ├── components/
    │   ├── ArticleDetail.tsx  # 文章详情页
    │   ├── ArticleList.tsx    # 文章列表 + 搜索
    │   ├── ArticleEditor.tsx  # 新建/编辑文章
    │   ├── Login.tsx          # 登录表单
    │   ├── Register.tsx       # 注册表单
    │   └── Profile.tsx        # 用户资料查看与修改
    └── hooks/
        └── useAuth.ts         # 简单的 JWT 管理（localStorage）
```

## 3. 与后端 API 的对应关系

### 3.1 配置与基础 URL

前端通过环境变量 `VITE_API_BASE_URL` 指定后端 API 根地址：

- 开发环境（本机）: 默认 `http://localhost:8080/api/v1`
- docker-compose 环境: `http://app:8080/api/v1`

`src/api/client.ts` 中创建 Axios 实例，并在请求拦截器中自动附加 `Authorization: Bearer <token>` 头，从 `localStorage.token` 读取。

### 3.2 用户认证

后端 API：

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- `GET /api/v1/users/profile`
- `PUT /api/v1/users/profile`

前端实现：

- `components/Login.tsx`：登录后存储 JWT，跳回 `/`
- `components/Register.tsx`：注册成功后自动登录
- `components/Profile.tsx`：展示并可更新用户名、邮箱、头像、简介

### 3.3 文章列表与搜索

后端 API：

- `GET /api/v1/articles`
  - 支持 `page`、`page_size`、`search` 等 Query 参数

前端实现：

- `components/ArticleList.tsx`
  - 使用 `useSearchParams` 读取 `page`/`page_size`/`search`
  - 请求：`GET /articles?page=...&page_size=...&search=...`
  - 展示分页、搜索框
  - 搜索框提交后更新 URL 的 `search` 参数，触发重新加载

### 3.4 文章详情

后端 API：

- `GET /api/v1/articles/:id`

前端实现：

- `components/ArticleDetail.tsx`
  - 从 `useParams()` 获取 `id`
  - 请求：`GET /articles/:id`
  - 展示标题、统计信息和正文内容
  - 登录用户可跳转到编辑页

### 3.5 新建 / 编辑文章

后端 API：

- `POST /api/v1/articles`
- `PUT /api/v1/articles/:id`

前端实现：

- `components/ArticleEditor.tsx`
  - 支持 `/articles/new` 与 `/articles/:id/edit`
  - 表单字段：标题、摘要、封面、状态、分类、标签、正文
  - 登录状态由 `useAuth` 校验，未登录用户会提示先登录

### 3.6 导航与菜单

- 顶部导航在登录后会显示 “New Article” 与 “Profile” 菜单
- 未登录状态下显示 Login / Register 入口

## 4. 鉴权与状态管理

本期实现采用极简方案：

- `hooks/useAuth.ts` 对 localStorage 中的 `token` 做封装，提供 `login/logout`
- Axios 拦截器在每次请求前读取 token，若存在则附加到请求头
- 表单组件（Profile、ArticleEditor 等）会在渲染阶段检查是否已登录
- 未使用 Redux / Zustand 等复杂方案，降低初期成本

后续如需扩展：

- 可引入 Context + Reducer 或 Zustand 管理用户信息与权限
- 为受保护的路由添加 `RequireAuth` 组件

## 5. Docker 与 docker-compose 集成

### 5.1 前端镜像构建

`frontend/Dockerfile`：

- 第一阶段：使用 `node:20-alpine` 安装依赖并执行 `npm run build`，生成静态文件 `dist/`
- 第二阶段：使用 `nginx:alpine` 作为运行时镜像，托管 `dist` 目录

### 5.2 docker-compose 集成

在项目根目录 `docker-compose.yml` 中新增了 `frontend` 服务：

```yaml
frontend:
  build: ./frontend
  container_name: enterprise-blog-frontend
  depends_on:
    - app
  environment:
    VITE_API_BASE_URL: "http://app:8080/api/v1"
  ports:
    - "3000:80"
```

说明：

- `depends_on: app` 确保后端服务先启动
- 构建阶段通过 `VITE_API_BASE_URL`（Build Args）写入 bundle，运行阶段 Nginx 以 `try_files` 支持前端路由
- 将容器的 `80` 端口映射到宿主机 `3000`

### 5.3 启动方式

在项目根目录：

```bash
docker-compose up -d postgres redis app frontend
```

然后访问：

- 前端：`http://localhost:3000`
- 后端 API：`http://localhost:8080/api/v1`

## 6. 开发与调试

### 6.1 本地开发

```bash
cd frontend
npm install
npm run dev
```

- 默认在 `http://localhost:3000` 启动 Vite Dev Server
- 确保后端在本机 `8080` 端口运行，或在 `frontend/.env` 配置 `VITE_API_BASE_URL`

### 6.2 构建与预览

```bash
cd frontend
npm run build
npm run preview
```

## 7. 后续优化方向

- 引入组件库（如 Ant Design / MUI）提升 UI 质量
- 增加更多页面：分类/标签列表、用户资料页、评论管理等
- 使用 Suspense 与代码分割优化性能
- 接入真实的监控与日志（如 Sentry）追踪前端错误


