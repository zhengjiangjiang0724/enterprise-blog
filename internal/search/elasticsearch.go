package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"enterprise-blog/internal/models"
	"enterprise-blog/pkg/logger"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
)

var esClient *elasticsearch.Client

const articleIndex = "articles"

// InitElasticsearch 初始化 Elasticsearch 客户端（可选，失败时仅记录日志）
func InitElasticsearch() {
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		// 未配置则不启用
		return
	}

	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		l := logger.GetLogger()
		l.Warn().Err(err).Msg("failed to init elasticsearch client, search disabled")
		return
	}

	// 简单 ping 校验
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := client.Info(client.Info.WithContext(ctx)); err != nil {
		l := logger.GetLogger()
		l.Warn().Err(err).Msg("failed to connect to elasticsearch, search disabled")
		return
	}

	esClient = client
	l := logger.GetLogger()
	l.Info().Str("url", url).Msg("Elasticsearch client initialized")
}

// IndexArticle 在 Elasticsearch 中索引一篇文章（用于创建/更新）
func IndexArticle(ctx context.Context, article *models.Article) error {
	if esClient == nil || article == nil {
		return nil
	}

	doc := map[string]interface{}{
		"title":        article.Title,
		"content":      article.Content,
		"excerpt":      article.Excerpt,
		"status":       string(article.Status),
		"author_id":    article.AuthorID.String(),
		"category_id":  nil,
		"published_at": article.PublishedAt,
		"created_at":   article.CreatedAt,
	}
	if article.CategoryID != nil {
		doc["category_id"] = article.CategoryID.String()
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esClient.Index.WithContext(ctx)
	res, err := esClient.Index(
		articleIndex,
		bytes.NewReader(body),
		req,
		esClient.Index.WithDocumentID(article.ID.String()),
		esClient.Index.WithRefresh("false"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch index error: %s", res.String())
	}
	return nil
}

// DeleteArticle 从 Elasticsearch 中删除文章文档
func DeleteArticle(ctx context.Context, id uuid.UUID) error {
	if esClient == nil {
		return nil
	}
	res, err := esClient.Delete(
		articleIndex,
		id.String(),
		esClient.Delete.WithContext(ctx),
		esClient.Delete.WithRefresh("false"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// 删除不存在的文档不视为错误
	return nil
}

// SearchArticles 使用 Elasticsearch 搜索文章，返回文章 ID 列表和总数
// query: 搜索关键词
// page: 页码
// pageSize: 每页数量
// filters: 可选的筛选条件（状态、分类ID、作者ID等）
// 返回: 文章ID列表、总数，如果搜索失败则返回错误
func SearchArticles(ctx context.Context, query string, page, pageSize int, filters ...map[string]interface{}) ([]uuid.UUID, int64, error) {
	if esClient == nil {
		return nil, 0, fmt.Errorf("elasticsearch not initialized")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	from := (page - 1) * pageSize

	// 构建查询
	esQuery := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  query,
			"fields": []string{"title^3", "excerpt^2", "content"},
		},
	}

	// 如果有筛选条件，使用bool查询
	if len(filters) > 0 && len(filters[0]) > 0 {
		mustClauses := []map[string]interface{}{esQuery}
		filterClauses := []map[string]interface{}{}

		filter := filters[0]
		if status, ok := filter["status"].(string); ok && status != "" {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"status": status},
			})
		}
		if categoryID, ok := filter["category_id"].(string); ok && categoryID != "" {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"category_id": categoryID},
			})
		}
		if authorID, ok := filter["author_id"].(string); ok && authorID != "" {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"author_id": authorID},
			})
		}

		if len(filterClauses) > 0 {
			esQuery = map[string]interface{}{
				"bool": map[string]interface{}{
					"must":   mustClauses,
					"filter": filterClauses,
				},
			}
		}
	}

	body := map[string]interface{}{
		"from": from,
		"size": pageSize,
		"query": esQuery,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}

	res, err := esClient.Search(
		esClient.Search.WithContext(ctx),
		esClient.Search.WithIndex(articleIndex),
		esClient.Search.WithBody(bytes.NewReader(reqBody)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	var ids []uuid.UUID
	for _, h := range result.Hits.Hits {
		if id, err := uuid.Parse(h.ID); err == nil {
			ids = append(ids, id)
		}
	}

	return ids, result.Hits.Total.Value, nil
}


