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

### 文章相关

#### 获取文章列表
```
GET /articles?page=1&page_size=10&status=published&category_id=xxx&tag_id=xxx&search=keyword&search_mode=es
```

**查询参数**:
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认10）
- `status`: 文章状态（draft/published/archived）
- `category_id`: 分类ID
- `tag_id`: 标签ID
- `search`: 搜索关键词
- `sort_by`: 排序字段（created_at/view_count等）
- `order`: 排序方向（asc/desc）
- `search_mode`: 可选，`es` 时使用 Elasticsearch 搜索；省略或其他值时使用 PostgreSQL 全文搜索

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
  "status": "draft",
  "category_id": "uuid",
  "tag_ids": ["uuid1", "uuid2"]
}
```

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

### 标签相关

#### 获取标签列表
```
GET /tags
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

