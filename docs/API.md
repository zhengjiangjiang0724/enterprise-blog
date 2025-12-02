# API文档

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`

## 认证

大多数API需要JWT认证，在请求头中携带Token：

```
Authorization: Bearer <token>
```

## API端点

### 认证相关

#### 用户注册
```
POST /auth/register
```

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "role": "reader"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "username": "testuser",
    "email": "test@example.com",
    "role": "reader",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 用户登录
```
POST /auth/login
```

**请求体**:
```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "jwt_token_here",
    "user": {
      "id": "uuid",
      "username": "testuser",
      "email": "test@example.com",
      "role": "reader"
    }
  }
}
```

#### 发送短信验证码
```
POST /auth/send-sms-code
```

**请求体**:
```json
{
  "phone": "13800138000"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "验证码已发送"
  }
}
```

**说明**:
- 验证码有效期为 5 分钟
- 同一手机号 1 分钟内只能发送一次验证码（防刷）
- 当前为模拟实现，验证码会在后端日志中输出（开发/测试环境）

#### 手机号验证码登录
```
POST /auth/login-phone
```

**请求体**:
```json
{
  "phone": "13800138000",
  "code": "123456"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "jwt_token_here",
    "user": {
      "id": "uuid",
      "username": "user_8000",
      "email": "13800138000@phone.local",
      "phone": "13800138000",
      "role": "reader"
    }
  }
}
```

**说明**:
- 如果手机号对应的用户不存在，系统会自动创建新用户
- 自动创建的用户默认角色为 `reader`，用户名为手机号后4位，邮箱为临时邮箱
- 验证码验证成功后会被标记为已使用，不能重复使用

### 用户相关

#### 获取当前用户信息
```
GET /users/profile
```
需要认证

#### 更新当前用户信息
```
PUT /users/profile
```
需要认证

#### 修改当前用户密码
```
PUT /users/password
```
需要认证

**请求体**:
```json
{
  "old_password": "当前密码",
  "new_password": "新密码（至少 6 位）"
}
```

**说明**:
- 后端会校验 `old_password` 是否正确，然后使用 bcrypt 重新哈希并更新存储。
- 修改成功后建议前端提示用户重新登录。

### 文章相关

#### 获取文章列表
```
GET /articles?page=1&page_size=10&status=published&category_id=xxx&tag_id=xxx&search=keyword&search_mode=es
```

**查询参数**:
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认10）
- `status`: 文章状态（draft/review/published/archived），公开列表通常只使用 `published`
- `category_id`: 分类ID
- `tag_id`: 标签ID
- `search`: 搜索关键词
- `sort_by`: 排序字段（created_at/view_count等）
- `order`: 排序方向（asc/desc）
- `search_mode`: 可选，`es` 时使用 Elasticsearch 搜索；省略或其他值时使用 PostgreSQL 全文搜索

说明：
- 不传 `status` 时，公开文章列表接口默认只返回 `published` 状态的文章。

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [...],
  "meta": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_page": 10
  }
}
```

#### 获取文章详情
```
GET /articles/:id
```

#### 文章点赞
```
POST /articles/:id/like
```

**说明**: 点赞一次将该文章的 `like_count` 加 1，目前不做去重控制（前端可根据需要做防重复点击）。

#### 通过Slug获取文章
```
GET /articles/slug/:slug
```

#### 创建文章
```
POST /articles
```
需要认证

**请求体**:
```json
{
  "title": "文章标题",
  "content": "文章内容",
  "excerpt": "文章摘要",
  "cover_image": "封面图片URL",
  "status": "draft",   // 可选：draft（草稿）/ review（提交审核）/ published（直接发布，需要有权限）
  "category_id": "uuid",
  "tag_ids": ["uuid1", "uuid2"]
}
```

说明：
- 普通作者通常通过“保存为草稿”或“提交审核”创建文章：
  - 草稿：`status = "draft"`，仅作者自己和管理员可在后台看到。
  - 提交审核：`status = "review"`，进入待审核队列，由管理员在后台审核后发布。
- 管理员可以直接创建 `published` 状态的文章。

#### 更新文章
```
PUT /articles/:id
```
需要认证

#### 删除文章
```
DELETE /articles/:id
```
需要认证

### 分类相关

#### 获取分类列表
```
GET /categories
```

#### 管理后台 - 分类管理

仅管理员可调用：

```
GET    /admin/categories        # 分类列表
POST   /admin/categories        # 新建分类
GET    /admin/categories/:id    # 分类详情
PUT    /admin/categories/:id    # 更新分类
DELETE /admin/categories/:id    # 删除分类
```

### 标签相关

#### 获取标签列表
```
GET /tags
```

#### 管理后台 - 标签管理

仅管理员可调用：

```
GET    /admin/tags        # 标签列表
POST   /admin/tags        # 新建标签
GET    /admin/tags/:id    # 标签详情
PUT    /admin/tags/:id    # 更新标签
DELETE /admin/tags/:id    # 删除标签
```

### 评论相关

#### 获取文章评论
```
GET /articles/:article_id/comments?page=1&page_size=20
```

#### 创建评论
```
POST /articles/:article_id/comments
```

**请求体**:
```json
{
  "parent_id": "uuid（可选，回复评论时使用）",
  "content": "评论内容",
  "author": "作者名称",
  "email": "author@example.com",
  "website": "https://example.com"
}
```

## 错误码

- `200`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `429`: 请求过于频繁
- `500`: 服务器内部错误

