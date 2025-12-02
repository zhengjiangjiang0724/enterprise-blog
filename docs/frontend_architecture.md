# 前端项目架构设计文档

## 1. 概述

前端项目是企业级博客系统的独立 Web 客户端，主要职责：

- 提供文章浏览、搜索、详情查看等功能
- 提供文章创建 / 编辑（Markdown 编辑与预览、草稿 / 审核 / 发布流程、封面图片选择）
- 提供评论、点赞、收藏等交互能力（实时数据更新）
- 提供图片上传和管理功能
- 提供用户登录 / 注册 / 个人资料管理以及后台用户管理入口
- 提供统一的返回按钮导航体验
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
    ├── App.tsx                 # 路由与页面布局，导航栏
    ├── main.tsx                # 入口文件，挂载 AuthProvider / MessageProvider
    ├── styles.css              # 全局样式（按钮、表单、文章卡片、评论区等）
    ├── api/
    │   ├── client.ts           # Axios 实例和拦截器
    │   └── types.ts            # 与后端对齐的基础类型（Article、User、Comment 等）
    ├── components/
    │   ├── ArticleDetail.tsx   # 文章详情（Markdown 渲染、点赞、收藏、评论、实时更新）
    │   ├── ArticleList.tsx     # 文章列表 + 搜索 + 分页 + 统计展示
    │   ├── ArticleEditor.tsx   # 新建/编辑文章（Markdown 编辑 & 预览、草稿 / 审核 / 发布、封面图片选择）
    │   ├── CommentSection.tsx  # 评论列表与发表评论表单（实时更新评论数）
    │   ├── Login.tsx           # 登录表单
    │   ├── Register.tsx        # 注册表单
    │   ├── Profile.tsx         # 用户资料查看与修改
    │   ├── ImageUpload.tsx     # 图片上传组件
    │   ├── ImageList.tsx       # 图片列表和搜索
    │   ├── ImagePicker.tsx     # 图片选择器（从图片库选择封面）
    │   ├── BackButton.tsx      # 返回按钮组件（统一导航）
    │   ├── AdminUserList.tsx   # 后台用户列表
    │   ├── AdminUserDetail.tsx # 后台用户详情
    │   ├── AdminArticleList.tsx   # 后台文章列表与筛选
    │   ├── AdminArticleDetail.tsx # 后台文章详情与状态流转（草稿 / 待审核 / 发布 / 归档）
    │   ├── AdminCategoryList.tsx  # 后台分类管理（增删改）
    │   ├── AdminTagList.tsx       # 后台标签管理（增删改）
    │   ├── Button.tsx          # 通用按钮组件（primary / secondary / danger / ghost）
    │   ├── MessageProvider.tsx # 全局消息提示（成功 / 错误）
    │   └── RouteGuards.tsx     # 路由守卫组件（RequireAuth / RequireRole）
    └── hooks/
        ├── useAuth.tsx         # 基于 Context 的 JWT 和用户信息管理
        └── useFavorites.ts     # 本地收藏文章管理（localStorage）
```

## 3. 与后端 API 的对应关系

### 3.1 配置与基础 URL

前端通过环境变量 `VITE_API_BASE_URL` 指定后端 API 根地址：

- 开发环境（本机）: 默认 `http://localhost:8080/api/v1`
- docker-compose 环境: `http://localhost:8080/api/v1`（通过 build args 写入）

`src/api/client.ts` 中创建 Axios 实例，并在请求拦截器中自动附加 `Authorization: Bearer <token>` 头，从 `localStorage.token` 读取。

### 3.2 用户认证与权限

后端 API：

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- `GET /api/v1/users/profile`
- `PUT /api/v1/users/profile`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/:id`

前端实现：

- `components/Login.tsx`：登录后存储 JWT，跳回文章列表 `/`
- `components/Register.tsx`：注册成功后自动登录
- `components/Profile.tsx`：展示并可更新用户名、邮箱、头像、简介
- `components/AdminUserList.tsx` / `AdminUserDetail.tsx`：仅管理员可访问的用户管理页面
- `components/RouteGuards.tsx`：`RequireAuth` / `RequireRole` 确保受保护页面只能在登录或具备特定角色时访问

### 3.3 文章列表与搜索

后端 API：

- `GET /api/v1/articles`
  - 支持 `page`、`page_size`、`search`、`status`、`category_id`、`sort_by`、`order` 等 Query 参数
  - `search` 参数使用 Elasticsearch 进行全文搜索（支持模糊匹配）

前端实现：

- `components/ArticleList.tsx`
  - 使用 `useSearchParams` 读取 `page`/`page_size`/`search`
  - 请求：`GET /articles?page=...&page_size=...&search=...`
  - 支持实时搜索（输入关键词后自动搜索）
  - 展示分页、搜索框
  - 搜索框提交后更新 URL 的 `search` 参数，触发重新加载

### 3.4 文章详情、点赞与评论

后端 API：

- `GET /api/v1/articles/:id`
- `POST /api/v1/articles/:id/like`
- `GET /api/v1/articles/:id/comments`
- `POST /api/v1/articles/:id/comments`

前端实现：

- `components/ArticleDetail.tsx`
  - 从 `useParams()` 获取 `id`
  - 请求：`GET /articles/:id`
  - 使用 `react-markdown` 渲染文章正文内容
  - 展示阅读数、点赞数、评论数（实时更新）
  - 提供“点赞”按钮（调用 `POST /articles/:id/like`，成功后立即更新点赞数）
  - 提供“收藏 / 已收藏”按钮（本地收藏，使用 `useFavorites` hook）
  - 登录用户可跳转到编辑页、删除文章
  - 集成 `BackButton` 组件，提供返回文章列表的导航
- `components/CommentSection.tsx`
  - 请求：`GET /articles/:id/comments` 分页加载评论
  - 请求：`POST /articles/:id/comments` 发表评论（登录用户自动带用户名 / 邮箱）
  - 评论提交成功后，立即更新评论数并重新加载评论列表（实时更新）
  - 通过 `onCommentAdded` 回调通知父组件更新文章评论数

### 3.5 新建 / 编辑文章（Markdown + 草稿箱 / 审核 / 发布流程）

后端 API：

- `POST /api/v1/articles`
- `PUT /api/v1/articles/:id`

前端实现：

- `components/ArticleEditor.tsx`
  - 支持 `/articles/new` 与 `/articles/:id/edit`
  - 表单字段：标题、摘要、封面、状态、正文内容、分类、标签
  - 使用 `react-markdown` 提供“编辑 / 预览”模式切换
  - 提供“保存为草稿”“提交审核”和“发布文章 / 更新并发布”三个按钮，通过 `status` 字段控制文章状态：
    - 草稿：`draft`，仅作者与管理员在后台可见
    - 待审核：`review`，仅管理员在后台审核列表中可见
    - 已发布：`published`，公开文章列表和详情可见
  - 封面图片支持两种方式：
    - 直接输入图片URL
    - 从图片库选择（集成 `ImagePicker` 组件）
  - 分类下拉选项与标签多选数据来源：
    - 分类：`GET /api/v1/categories`
    - 标签：`GET /api/v1/tags`
  - 提交文章时会携带：
    - `category_id`：选中的分类 ID（可为空）
    - `tag_ids`：选中的标签 ID 列表（可为空）
  - 登录状态由 `RequireAuth` / `useAuth` 共同校验，未登录用户会被重定向到登录页
  - 集成 `BackButton` 组件，提供返回导航

### 3.6 图片管理

后端 API：

- `POST /api/v1/images/upload` - 上传图片（需认证）
- `GET /api/v1/images` - 获取图片列表（支持分页、搜索、标签筛选）
- `GET /api/v1/images/:id` - 获取图片详情
- `PUT /api/v1/images/:id` - 更新图片信息（需认证）
- `DELETE /api/v1/images/:id` - 删除图片（需认证）
- `GET /uploads/images/:filename` - 访问图片文件（静态文件服务）

前端实现：

- `components/ImageUpload.tsx`
  - 图片上传表单（支持文件选择、描述、标签）
  - 文件类型和大小验证
  - 上传成功后显示图片预览
- `components/ImageList.tsx`
  - 图片列表展示（网格布局）
  - 支持搜索、标签筛选、分页
  - 图片预览和详情查看
  - 图片删除功能（需认证）
- `components/ImagePicker.tsx`
  - 图片选择器弹窗
  - 从图片库中选择图片作为文章封面
  - 支持搜索和分页
  - 选择后自动填充封面URL

### 3.7 返回按钮导航

前端实现：

- `components/BackButton.tsx`
  - 通用的返回按钮组件
  - 支持自定义返回路径和标签文本
  - 集成到以下页面：
    - `ArticleDetail` - 返回文章列表
    - `ArticleEditor` - 返回文章列表或详情
    - `Profile` - 返回首页
    - `ImageUpload` / `ImageList` - 返回首页
    - `AdminUserDetail` / `AdminArticleDetail` - 返回对应的列表页

### 3.8 实时数据更新

前端实现：

- **点赞实时更新**：
  - `ArticleDetail` 组件中，点赞成功后立即更新本地状态中的 `like_count`
  - 无需刷新页面即可看到最新的点赞数
- **收藏实时更新**：
  - 使用 `useFavorites` hook 管理本地收藏状态
  - 收藏/取消收藏后立即更新UI
- **评论实时更新**：
  - `CommentSection` 组件中，评论提交成功后：
    - 立即重新加载评论列表
    - 通过 `onCommentAdded` 回调通知父组件更新 `comment_count`
  - `ArticleDetail` 组件接收回调后立即更新本地状态中的 `comment_count`

### 3.9 导航与菜单

- 顶部导航在登录后会显示“新建文章”“个人资料”“图片管理”“用户管理（管理员）”等菜单
- 未登录状态下显示 Login / Register 入口
- 右上角展示当前登录用户的用户名与角色，并提供“退出登录”按钮

## 4. 鉴权与状态管理

当前实现：

- `hooks/useAuth.tsx` 使用 React Context 管理全局 `token` 和 `user` 状态，并持久化到 `localStorage`
- Axios 拦截器在每次请求前读取 token，若存在则附加到请求头
- `components/RouteGuards.tsx` 提供 `RequireAuth` / `RequireRole` 路由守卫，保护文章编辑、个人资料、后台用户管理等页面
- `components/MessageProvider.tsx` 提供全局消息提示，用于登录 / 注册 / 文章操作 / 资料更新等场景的统一成功/失败提示

后续如需扩展：

- 可引入 Zustand / Redux 等更强大的状态管理库
- 在 Axios 响应拦截器中统一处理 401/403，自动登出并重定向到登录页

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

## 7. 图片URL处理

前端需要正确处理图片URL：

- 图片URL格式：`/uploads/images/{filename}`
- 完整访问地址：`http://localhost:8080/uploads/images/{filename}`
- 前端通过 `getImageUrl` 函数处理相对路径：
  ```typescript
  const getImageUrl = (url: string): string => {
    if (url.startsWith("http")) {
      return url;
    }
    const baseURL = import.meta.env.VITE_API_BASE_URL?.replace("/api/v1", "") || "http://localhost:8080";
    return `${baseURL}${url}`;
  };
  ```
- 在 `ArticleDetail` 和 `ArticleEditor` 中，封面图片URL会自动处理

## 8. 后续优化方向

- 引入组件库（如 Ant Design / MUI）提升 UI 质量
- 增加更多页面：分类/标签列表、用户资料页、评论管理等
- 使用 Suspense 与代码分割优化性能
- 接入真实的监控与日志（如 Sentry）追踪前端错误
- 图片缩略图生成和CDN集成
- WebSocket支持实时通知（点赞、评论等）


