package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"
	"enterprise-blog/internal/search"
	"enterprise-blog/pkg/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ArticleService struct {
	articleRepo  *repository.ArticleRepository
	categoryRepo *repository.CategoryRepository
	tagRepo      *repository.TagRepository
}

func NewArticleService(
	articleRepo *repository.ArticleRepository,
	categoryRepo *repository.CategoryRepository,
	tagRepo *repository.TagRepository,
) *ArticleService {
	return &ArticleService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
}

func (s *ArticleService) Create(authorID uuid.UUID, req *models.ArticleCreate) (*models.Article, error) {
	// 生成slug
	slug := GenerateSlug(req.Title)
	if slug == "" {
		slug = "article"
	}
	// 检查slug是否已存在，如果存在则添加数字后缀
	originalSlug := slug
	counter := 1

	// 生成摘要
	excerpt := req.Excerpt
	if excerpt == "" && len(req.Content) > 200 {
		excerpt = req.Content[:200] + "..."
	} else if excerpt == "" {
		excerpt = req.Content
	}

	article := &models.Article{
		Title:      req.Title,
		Slug:       slug,
		Content:    req.Content,
		Excerpt:    excerpt,
		CoverImage: req.CoverImage,
		Status:     req.Status,
		AuthorID:   authorID,
	}

	if article.Status == "" {
		article.Status = models.StatusDraft
	}

	// 创建时如果遇到 slug 唯一约束冲突，则自动追加数字后缀重试几次
	const maxSlugRetries = 5
	for retries := 0; retries < maxSlugRetries; retries++ {
		article.Slug = slug

		if err := s.articleRepo.Create(context.Background(), article); err != nil {
			// 唯一约束冲突：尝试下一个 slug
			if isSlugUniqueViolation(err) {
				slug = fmt.Sprintf("%s-%d", originalSlug, counter)
				counter++
				continue
			}
			return nil, fmt.Errorf("failed to create article: %w", err)
		}

		// 创建成功，重新从数据库获取完整数据（含作者等关联）
		created, err := s.articleRepo.GetByIDWithContext(context.Background(), article.ID)
		if err == nil {
			// 写入详情缓存，并清理列表缓存
			_ = cacheArticleDetail(created)
			clearArticleListCache()

			// 异步同步到 Elasticsearch（如果已启用）
			go func(a *models.Article) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_ = search.IndexArticle(ctx, a)
			}(created)

			return created, nil
		}
		return created, err
	}

	return nil, fmt.Errorf("failed to create article: slug already exists after %d retries", maxSlugRetries)
}

func (s *ArticleService) GetByID(id uuid.UUID) (*models.Article, error) {
	// 优先从缓存读取
	if article, err := getArticleDetailFromCache(id); err == nil && article != nil {
		// 增加浏览计数（缓冲）
		go incrementArticleViewCountBuffered(article.ID)
		return article, nil
	}

	article, err := s.articleRepo.GetByIDWithContext(context.Background(), id)
	if err != nil {
		return nil, err
	}

	// 写入缓存（忽略错误）
	_ = cacheArticleDetail(article)

	// 增加浏览计数：优先写入 Redis 作为缓冲，失败时退回到数据库自增
	go incrementArticleViewCountBuffered(article.ID)

	return article, nil
}

func (s *ArticleService) GetBySlug(slug string) (*models.Article, error) {
	article, err := s.articleRepo.GetBySlugWithContext(context.Background(), slug)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	go s.articleRepo.IncrementViewCount(article.ID)

	return article, nil
}

func (s *ArticleService) Update(id uuid.UUID, req *models.ArticleUpdate) (*models.Article, error) {
	article, err := s.articleRepo.GetByIDWithContext(context.Background(), id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		article.Title = *req.Title
		// 如果标题改变，更新slug
		article.Slug = GenerateSlug(*req.Title)
	}

	if req.Content != nil {
		article.Content = *req.Content
		// 如果内容改变但没有摘要，自动生成摘要
		if req.Excerpt == nil {
			if len(*req.Content) > 200 {
				article.Excerpt = (*req.Content)[:200] + "..."
			} else {
				article.Excerpt = *req.Content
			}
		}
	}

	if req.Excerpt != nil {
		article.Excerpt = *req.Excerpt
	}

	if req.CoverImage != nil {
		article.CoverImage = *req.CoverImage
	}

	if req.Status != nil {
		article.Status = *req.Status
	}

	// 简化：不再更新分类和标签，文章仅保留基本信息

	if err := s.articleRepo.Update(article); err != nil {
		return nil, err
	}

	updated, err := s.articleRepo.GetByIDWithContext(context.Background(), id)
	if err == nil {
		// 更新详情缓存，并清理列表缓存
		_ = cacheArticleDetail(updated)
		clearArticleListCache()

		// 异步同步到 Elasticsearch
		go func(a *models.Article) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = search.IndexArticle(ctx, a)
		}(updated)
	}

	return updated, err
}

func (s *ArticleService) Delete(id uuid.UUID) error {
	if err := s.articleRepo.Delete(id); err != nil {
		return err
	}
	// 删除详情缓存并清理列表缓存
	deleteArticleDetailCache(id)
	clearArticleListCache()

	// 异步从 Elasticsearch 删除文档
	go func(articleID uuid.UUID) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = search.DeleteArticle(ctx, articleID)
	}(id)

	return nil
}

func (s *ArticleService) List(query models.ArticleQuery) ([]*models.Article, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 尝试从缓存读取列表
	if articles, total, err := getArticleListFromCache(query); err == nil && articles != nil {
		return articles, total, nil
	}

	ctx := context.Background()
	articles, total, err := s.articleRepo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存（忽略错误）
	_ = cacheArticleList(query, articles, total)

	return articles, total, nil
}

func (s *ArticleService) Like(id uuid.UUID) error {
	// 点赞计数：优先写入 Redis 作为缓冲，失败时退回到数据库自增
	if err := incrementArticleLikeCountBuffered(id); err != nil {
		// 记录日志，但不中断请求
		return s.articleRepo.IncrementLikeCount(id)
	}
	return nil
}

// SearchWithElasticsearch 使用 Elasticsearch 搜索文章，并回到数据库取完整数据
func (s *ArticleService) SearchWithElasticsearch(query string, page, pageSize int) ([]*models.Article, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ids, total, err := search.SearchArticles(ctx, query, page, pageSize)
	if err != nil {
		// 如果 ES 搜索失败，则回退到 PostgreSQL 全文搜索，保证接口可用
		l := logger.GetLogger()
		l.Warn().Err(err).Msg("Elasticsearch search failed, fallback to PostgreSQL fulltext search")

		fallbackQuery := models.ArticleQuery{
			Page:     page,
			PageSize: pageSize,
			Search:   query,
		}
		return s.List(fallbackQuery)
	}
	if len(ids) == 0 {
		return []*models.Article{}, 0, nil
	}

	// 简单实现：逐个从数据库读取，后续可优化为批量查询
	articles := make([]*models.Article, 0, len(ids))
	for _, id := range ids {
		art, err := s.articleRepo.GetByIDWithContext(ctx, id)
		if err != nil {
			continue
		}
		articles = append(articles, art)
	}
	return articles, total, nil
}

// isSlugUniqueViolation 判断是否为 articles.slug 唯一约束冲突
func isSlugUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// 目前使用字符串包含判断，兼容 pq/pgx 等驱动的错误文案
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint") &&
		strings.Contains(msg, "articles_slug_key")
}

const (
	redisArticleViewKeyPrefix = "blog:article:view:"
	redisArticleLikeKeyPrefix = "blog:article:like:"
	redisArticleDetailPrefix  = "blog:article:detail:"
	redisArticleListPrefix    = "blog:article:list:"
)

// incrementArticleViewCountBuffered 将浏览计数写入 Redis，失败时退回到数据库
func incrementArticleViewCountBuffered(id uuid.UUID) {
	if database.RedisClient == nil {
		_ = (&repository.ArticleRepository{}).IncrementViewCount(id)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	key := redisArticleViewKeyPrefix + id.String()
	if err := database.RedisClient.Incr(ctx, key).Err(); err != nil {
		l := logger.GetLogger()
		l.Warn().Err(err).Str("key", key).Msg("failed to increment view count in redis, fallback to db")
		_ = (&repository.ArticleRepository{}).IncrementViewCount(id)
	}
}

// incrementArticleLikeCountBuffered 将点赞计数写入 Redis，失败时返回错误由上层回退处理
func incrementArticleLikeCountBuffered(id uuid.UUID) error {
	if database.RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	key := redisArticleLikeKeyPrefix + id.String()
	if err := database.RedisClient.Incr(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

// FlushArticleCountersFromRedis 将 Redis 中的浏览 / 点赞增量批量回刷到数据库
func FlushArticleCountersFromRedis(ctx context.Context) error {
	if database.RedisClient == nil {
		return nil
	}

	l := logger.GetLogger()
	rdb := database.RedisClient

	// 浏览计数
	if err := flushCounterPrefix(ctx, rdb, redisArticleViewKeyPrefix, func(id uuid.UUID, delta int64) error {
		query := `UPDATE articles SET view_count = view_count + $1 WHERE id = $2`
		return database.DB.Exec(query, delta, id).Error
	}); err != nil {
		l.Error().Err(err).Msg("failed to flush view counters from redis")
	}

	// 点赞计数
	if err := flushCounterPrefix(ctx, rdb, redisArticleLikeKeyPrefix, func(id uuid.UUID, delta int64) error {
		query := `UPDATE articles SET like_count = like_count + $1 WHERE id = $2`
		return database.DB.Exec(query, delta, id).Error
	}); err != nil {
		l.Error().Err(err).Msg("failed to flush like counters from redis")
	}

	return nil
}

// flushCounterPrefix 扫描指定前缀的计数键，读取增量并应用到数据库，然后删除键
func flushCounterPrefix(
	ctx context.Context,
	rdb redisCmdable,
	prefix string,
	apply func(id uuid.UUID, delta int64) error,
) error {
	var cursor uint64
	for {
		keys, next, err := rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		cursor = next
		for _, key := range keys {
			val, err := rdb.Get(ctx, key).Int64()
			if err != nil {
				continue
			}
			if val == 0 {
				_ = rdb.Del(ctx, key).Err()
				continue
			}
			idStr := strings.TrimPrefix(key, prefix)
			id, err := uuid.Parse(idStr)
			if err != nil {
				_ = rdb.Del(ctx, key).Err()
				continue
			}
			if err := apply(id, val); err != nil {
				// 若更新失败，保留键以便下次重试
				continue
			}
			_ = rdb.Del(ctx, key).Err()
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

// redisCmdable 抽象 go-redis 客户端用于测试
type redisCmdable interface {
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// ---- 文章详情 / 列表缓存 ----

func getArticleDetailFromCache(id uuid.UUID) (*models.Article, error) {
	if database.RedisClient == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	key := redisArticleDetailPrefix + id.String()
	val, err := database.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var article models.Article
	if err := json.Unmarshal(val, &article); err != nil {
		return nil, err
	}
	return &article, nil
}

func cacheArticleDetail(article *models.Article) error {
	if database.RedisClient == nil || article == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	key := redisArticleDetailPrefix + article.ID.String()
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return database.RedisClient.Set(ctx, key, data, 60*time.Second).Err()
}

func deleteArticleDetailCache(id uuid.UUID) {
	if database.RedisClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	key := redisArticleDetailPrefix + id.String()
	_ = database.RedisClient.Del(ctx, key).Err()
}

type cachedArticleList struct {
	Articles []*models.Article `json:"articles"`
	Total    int64             `json:"total"`
}

func buildArticleListCacheKey(q models.ArticleQuery) string {
	var b strings.Builder
	b.WriteString(redisArticleListPrefix)
	b.WriteString(fmt.Sprintf("p=%d&ps=%d", q.Page, q.PageSize))
	if q.Status != "" {
		b.WriteString("&status=")
		b.WriteString(string(q.Status))
	}
	if q.Search != "" {
		b.WriteString("&search=")
		b.WriteString(q.Search)
	}
	if q.SortBy != "" {
		b.WriteString("&sort=")
		b.WriteString(q.SortBy)
	}
	if q.Order != "" {
		b.WriteString("&order=")
		b.WriteString(q.Order)
	}
	if q.CategoryID != nil {
		b.WriteString("&cat=")
		b.WriteString(q.CategoryID.String())
	}
	if q.TagID != nil {
		b.WriteString("&tag=")
		b.WriteString(q.TagID.String())
	}
	if q.AuthorID != nil {
		b.WriteString("&author=")
		b.WriteString(q.AuthorID.String())
	}
	return b.String()
}

func getArticleListFromCache(q models.ArticleQuery) ([]*models.Article, int64, error) {
	if database.RedisClient == nil {
		return nil, 0, fmt.Errorf("redis not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	key := buildArticleListCacheKey(q)
	val, err := database.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, 0, err
	}
	var cached cachedArticleList
	if err := json.Unmarshal(val, &cached); err != nil {
		return nil, 0, err
	}
	return cached.Articles, cached.Total, nil
}

func cacheArticleList(q models.ArticleQuery, articles []*models.Article, total int64) error {
	if database.RedisClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	key := buildArticleListCacheKey(q)
	data, err := json.Marshal(cachedArticleList{
		Articles: articles,
		Total:    total,
	})
	if err != nil {
		return err
	}
	// 列表数据可以稍长一点 TTL
	return database.RedisClient.Set(ctx, key, data, 120*time.Second).Err()
}

// clearArticleListCache 简单粗暴地清理所有文章列表缓存（数据更新后调用）
func clearArticleListCache() {
	if database.RedisClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	var cursor uint64
	for {
		keys, next, err := database.RedisClient.Scan(ctx, cursor, redisArticleListPrefix+"*", 100).Result()
		if err != nil {
			return
		}
		cursor = next
		if len(keys) > 0 {
			_ = database.RedisClient.Del(ctx, keys...).Err()
		}
		if cursor == 0 {
			break
		}
	}
}

