// Package search 提供全文搜索功能，基于 Elasticsearch 实现
//
// 设计思路：
// 1. 使用 Elasticsearch 作为全文搜索引擎，提供高性能的搜索能力
// 2. 支持多种搜索策略：精确匹配、前缀匹配、模糊匹配、通配符匹配
// 3. 支持多字段搜索，不同字段设置不同权重（title > excerpt > content）
// 4. 支持筛选和排序功能
// 5. 优雅降级：如果 Elasticsearch 未启用或连接失败，不影响系统正常运行
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"enterprise-blog/internal/config"
	"enterprise-blog/internal/models"
	"enterprise-blog/pkg/logger"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
)

// esClient Elasticsearch 客户端实例（全局变量，单例模式）
// 在 InitElasticsearch 中初始化，如果初始化失败则为 nil
var esClient *elasticsearch.Client

// articleIndex Elasticsearch 索引名称，用于存储文章数据
const articleIndex = "articles"

// InitElasticsearch 初始化 Elasticsearch 客户端
//
// 功能说明：
// - 从配置文件读取 Elasticsearch 配置（URL、是否启用）
// - 创建 Elasticsearch 客户端连接
// - 执行健康检查（ping），确保连接可用
// - 如果初始化失败，仅记录警告日志，不影响系统启动
//
// 设计考虑：
// - 使用优雅降级策略：Elasticsearch 不可用时，系统仍可正常运行
// - 使用超时控制（2秒），避免启动时长时间阻塞
// - 单例模式：全局只有一个客户端实例，节省资源
//
// 面试要点：
// - 为什么使用单例模式？避免重复创建连接，节省资源
// - 为什么使用优雅降级？提高系统可用性，即使搜索服务不可用，其他功能仍可用
func InitElasticsearch() {
	if !config.AppConfig.Elasticsearch.Enabled {
		// Elasticsearch 未启用
		return
	}

	url := config.AppConfig.Elasticsearch.URL
	if url == "" {
		// 未配置URL则不启用
		l := logger.GetLogger()
		l.Warn().Msg("Elasticsearch enabled but URL not configured, search disabled")
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
//
// 参数说明：
// - ctx: 上下文，用于控制请求超时和取消
// - article: 要索引的文章对象
//
// 返回值：
// - error: 如果索引失败则返回错误，否则返回 nil
//
// 功能说明：
// - 将文章数据转换为 Elasticsearch 文档格式
// - 使用文章 ID 作为文档 ID，确保同一篇文章的更新会覆盖旧文档
// - 设置 refresh="false"，不立即刷新索引，提高性能（最终一致性）
//
// 数据结构：
// - title: 文章标题（用于搜索）
// - content: 文章内容（用于搜索）
// - excerpt: 文章摘要（用于搜索）
// - status: 文章状态（用于筛选）
// - author_id: 作者ID（用于筛选）
// - category_id: 分类ID（用于筛选，可为空）
// - published_at: 发布时间（用于排序）
// - created_at: 创建时间（用于排序）
//
// 设计考虑：
// - 使用 UUID 作为文档 ID，确保全局唯一性
// - 异步索引：通常在 Service 层异步调用，不阻塞主流程
// - 错误处理：索引失败不影响数据库操作，仅记录日志
//
// 面试要点：
// - 为什么使用 refresh="false"？提高写入性能，但会有短暂延迟（最终一致性）
// - 如何处理索引失败？记录日志，不影响主流程，可以后续重试
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
//
// 参数说明：
// - ctx: 上下文，用于控制请求超时和取消
// - id: 要删除的文章 UUID
//
// 返回值：
// - error: 如果删除失败则返回错误，否则返回 nil
//
// 功能说明：
// - 根据文章 ID 删除 Elasticsearch 中的文档
// - 如果文档不存在，不视为错误（幂等性）
// - 设置 refresh="false"，不立即刷新索引
//
// 设计考虑：
// - 幂等性：多次删除同一文档不会报错
// - 软删除：如果数据库使用软删除，这里也应该同步删除，避免搜索结果中出现已删除文章
//
// 面试要点：
// - 为什么删除不存在的文档不报错？保证幂等性，简化错误处理
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
//
// 参数说明：
// - ctx: 上下文，用于控制请求超时和取消
// - query: 文章查询条件，包含：
//   - Search: 搜索关键词（可选）
//   - Status: 文章状态筛选（可选）
//   - CategoryID: 分类ID筛选（可选）
//   - AuthorID: 作者ID筛选（可选）
//   - Page: 页码（默认1）
//   - PageSize: 每页数量（默认10，最大50）
//   - SortBy: 排序字段（可选）
//   - Order: 排序方向（asc/desc，可选）
//
// 返回值：
// - []uuid.UUID: 匹配的文章ID列表（按排序顺序）
// - int64: 匹配的文章总数（用于分页）
// - error: 如果搜索失败则返回错误
//
// 搜索策略（多策略组合，提高搜索准确性和召回率）：
// 1. 精确匹配（multi_match, best_fields）：在标题、摘要、内容中精确匹配，权重最高
//   - title 权重 5，excerpt 权重 3，content 权重 1
//
// 2. 前缀匹配（multi_match, phrase_prefix）：支持部分词匹配，如输入"Go"可以匹配"Golang"
// 3. 模糊匹配（match, fuzziness: AUTO）：支持拼写错误，如"golang"可以匹配"golang"
// 4. 通配符匹配（query_string）：支持通配符，如"*go*"可以匹配包含"go"的所有词
//
// 设计考虑：
// - 使用 bool 查询的 should 子句，至少匹配一个条件即可
// - 不同匹配策略有不同的权重，精确匹配优先级最高
// - 转义特殊字符，防止查询注入攻击
// - 支持筛选条件（status、category、author），使用 filter 子句（不计算相关性分数，性能更好）
// - 默认按创建时间倒序排序，最新的在前
// - 分页参数验证和限制，防止恶意请求
//
// 性能优化：
// - filter 子句不计算相关性分数，比 must 子句性能更好
// - 使用白名单验证排序字段，防止注入攻击
// - 限制每页最大数量（50），防止单次查询返回过多数据
//
// 错误处理：
// - 解析 Elasticsearch 错误响应，提取详细的错误信息
// - 支持 root_cause 错误信息，便于调试
//
// 面试要点：
// - 为什么使用多种搜索策略？提高搜索的准确性和召回率，满足不同用户需求
// - 为什么使用 filter 而不是 must？filter 不计算相关性分数，性能更好，适合精确匹配
// - 如何处理特殊字符？转义特殊字符，防止查询注入和查询错误
// - 如何保证搜索性能？使用 filter、限制分页大小、合理使用索引
func SearchArticles(ctx context.Context, query models.ArticleQuery) ([]uuid.UUID, int64, error) {
	if esClient == nil {
		return nil, 0, fmt.Errorf("elasticsearch not initialized")
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 50 {
		query.PageSize = 50
	}

	from := (query.Page - 1) * query.PageSize

	// 构建查询条件
	var esQuery map[string]interface{}

	if query.Search != "" {
		// 转义搜索关键词中的特殊字符，防止查询注入和查询错误
		// Elasticsearch 的 query_string 查询支持特殊字符，需要转义以避免被解释为操作符
		// 转义的字符包括：* ? + - && || ! ( ) { } [ ] ^ " ~ :
		escapedSearch := strings.ReplaceAll(query.Search, "*", "\\*")
		escapedSearch = strings.ReplaceAll(escapedSearch, "?", "\\?")
		escapedSearch = strings.ReplaceAll(escapedSearch, "+", "\\+")
		escapedSearch = strings.ReplaceAll(escapedSearch, "-", "\\-")
		escapedSearch = strings.ReplaceAll(escapedSearch, "&&", "\\&&")
		escapedSearch = strings.ReplaceAll(escapedSearch, "||", "\\||")
		escapedSearch = strings.ReplaceAll(escapedSearch, "!", "\\!")
		escapedSearch = strings.ReplaceAll(escapedSearch, "(", "\\(")
		escapedSearch = strings.ReplaceAll(escapedSearch, ")", "\\)")
		escapedSearch = strings.ReplaceAll(escapedSearch, "{", "\\{")
		escapedSearch = strings.ReplaceAll(escapedSearch, "}", "\\}")
		escapedSearch = strings.ReplaceAll(escapedSearch, "[", "\\[")
		escapedSearch = strings.ReplaceAll(escapedSearch, "]", "\\]")
		escapedSearch = strings.ReplaceAll(escapedSearch, "^", "\\^")
		escapedSearch = strings.ReplaceAll(escapedSearch, "\"", "\\\"")
		escapedSearch = strings.ReplaceAll(escapedSearch, "~", "\\~")
		escapedSearch = strings.ReplaceAll(escapedSearch, ":", "\\:")

		// 使用 bool 查询组合多种搜索策略，提高搜索准确性和召回率
		// should 子句：至少匹配一个条件即可（minimum_should_match: 1）
		// 不同策略有不同的权重和优先级，精确匹配权重最高
		shouldClauses := []map[string]interface{}{
			// 策略1：精确匹配（高优先级）
			// multi_match 的 best_fields 类型：在多个字段中查找，返回最佳匹配字段的分数
			// 字段权重：title^5 表示标题权重是内容的5倍，excerpt^3 表示摘要权重是内容的3倍
			// 这样标题匹配的文章会排在前面
			{
				"multi_match": map[string]interface{}{
					"query":  query.Search,
					"fields": []string{"title^5", "excerpt^3", "content"},
					"type":   "best_fields",
				},
			},
			// 策略2：前缀匹配（支持部分匹配）
			// phrase_prefix 类型：支持短语前缀匹配，如输入"Go"可以匹配"Golang"
			// 适用于用户输入不完整的情况
			{
				"multi_match": map[string]interface{}{
					"query":  query.Search,
					"fields": []string{"title^3", "excerpt^2", "content"},
					"type":   "phrase_prefix",
				},
			},
			// 策略3：模糊匹配（支持拼写错误）
			// fuzziness: "AUTO" 表示自动计算编辑距离，支持拼写错误纠正
			// 如"golang"可以匹配"golang"（即使有拼写错误）
			// boost: 2.0 表示标题的模糊匹配权重是内容的2倍
			{
				"match": map[string]interface{}{
					"title": map[string]interface{}{
						"query":     query.Search,
						"fuzziness": "AUTO",
						"boost":     2.0,
					},
				},
			},
			{
				"match": map[string]interface{}{
					"content": map[string]interface{}{
						"query":     query.Search,
						"fuzziness": "AUTO",
					},
				},
			},
			// 策略4：通配符匹配（支持部分词匹配）
			// query_string 查询：支持通配符 * 和 ?，如 "*go*" 可以匹配包含 "go" 的所有词
			// lenient: true 表示允许查询错误，不会导致整个查询失败（容错性）
			// default_operator: "OR" 表示多个词之间是或的关系
			{
				"query_string": map[string]interface{}{
					"query":            fmt.Sprintf("*%s*", escapedSearch),
					"fields":           []string{"title^2", "excerpt", "content"},
					"default_operator": "OR",
					"lenient":          true, // 允许查询错误，不会导致整个查询失败
				},
			},
		}

		// 构建 bool 查询
		// bool 查询是 Elasticsearch 最强大的查询类型，支持组合多个查询条件
		// should: 应该匹配的条件（至少匹配 minimum_should_match 个）
		// filter: 必须匹配的条件（不计算相关性分数，性能更好）
		boolQuery := map[string]interface{}{
			"should":               shouldClauses,
			"minimum_should_match": 1, // 至少匹配一个 should 条件即可
		}

		// 添加筛选条件（使用 filter 子句，不计算相关性分数，性能更好）
		// filter 子句用于精确匹配，如状态、分类、作者等
		filterClauses := []map[string]interface{}{}

		// 状态筛选
		if query.Status != "" {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"status": string(query.Status)},
			})
		}

		// 分类筛选
		if query.CategoryID != nil {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"category_id": query.CategoryID.String()},
			})
		}

		// 作者筛选
		if query.AuthorID != nil {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"author_id": query.AuthorID.String()},
			})
		}

		// 如果有筛选条件，添加到 bool 查询中
		if len(filterClauses) > 0 {
			boolQuery["filter"] = filterClauses
		}

		esQuery = map[string]interface{}{
			"bool": boolQuery,
		}
	} else {
		// 如果没有搜索关键词，只使用筛选条件
		filterClauses := []map[string]interface{}{}

		if query.Status != "" {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"status": string(query.Status)},
			})
		}
		if query.CategoryID != nil {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"category_id": query.CategoryID.String()},
			})
		}
		if query.AuthorID != nil {
			filterClauses = append(filterClauses, map[string]interface{}{
				"term": map[string]interface{}{"author_id": query.AuthorID.String()},
			})
		}

		if len(filterClauses) > 0 {
			esQuery = map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": filterClauses,
				},
			}
		} else {
			// 如果没有筛选条件，匹配所有文档
			esQuery = map[string]interface{}{
				"match_all": map[string]interface{}{},
			}
		}
	}

	// 构建排序：默认按创建时间倒序（最新的在前）
	// 这是最常见的排序需求，用户通常想看最新的文章
	sort := []map[string]interface{}{
		{
			"created_at": map[string]interface{}{
				"order": "desc",
			},
		},
	}

	// 如果指定了其他排序字段，优先使用指定的排序
	// 支持自定义排序字段和排序方向，提高灵活性
	if query.SortBy != "" {
		order := "desc" // 默认倒序
		if query.Order != "" {
			order = strings.ToLower(query.Order)
		}
		// 验证排序字段（白名单验证，防止注入攻击）
		// 只允许排序指定的字段，防止用户通过排序字段进行注入攻击
		allowedSortFields := map[string]string{
			"created_at":   "created_at",
			"updated_at":   "updated_at",
			"published_at": "published_at",
		}
		if sortField, ok := allowedSortFields[query.SortBy]; ok {
			sort = []map[string]interface{}{
				{
					sortField: map[string]interface{}{
						"order": order,
					},
				},
			}
		}
	}

	body := map[string]interface{}{
		"from":  from,
		"size":  query.PageSize,
		"query": esQuery,
		"sort":  sort,
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
		return nil, 0, fmt.Errorf("elasticsearch request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 读取错误响应体，提取详细的错误信息
		// Elasticsearch 的错误响应格式：
		// {
		//   "error": {
		//     "type": "query_shard_exception",
		//     "reason": "详细错误原因",
		//     "root_cause": [{"reason": "根本原因"}]
		//   }
		// }
		var errBody map[string]interface{}
		if decodeErr := json.NewDecoder(res.Body).Decode(&errBody); decodeErr == nil {
			if errorInfo, ok := errBody["error"].(map[string]interface{}); ok {
				// 优先使用 reason 字段的错误信息
				if reason, ok := errorInfo["reason"].(string); ok {
					return nil, 0, fmt.Errorf("elasticsearch search error: %s", reason)
				}
				// 如果没有 reason，尝试从 root_cause 获取
				if rootCause, ok := errorInfo["root_cause"].([]interface{}); ok && len(rootCause) > 0 {
					if firstCause, ok := rootCause[0].(map[string]interface{}); ok {
						if reason, ok := firstCause["reason"].(string); ok {
							return nil, 0, fmt.Errorf("elasticsearch search error: %s", reason)
						}
					}
				}
			}
		}
		// 如果无法解析错误响应，返回原始错误信息（包含查询内容，便于调试）
		bodyBytes, _ := json.Marshal(reqBody)
		return nil, 0, fmt.Errorf("elasticsearch search error (status: %d): query=%s", res.StatusCode, string(bodyBytes))
	}

	// 解析搜索结果
	// Elasticsearch 的搜索结果格式：
	// {
	//   "hits": {
	//     "total": {"value": 100},  // 匹配的总数
	//     "hits": [                 // 匹配的文档列表
	//       {"_id": "uuid1"},
	//       {"_id": "uuid2"}
	//     ]
	//   }
	// }
	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"` // 匹配的总数（用于分页）
			} `json:"total"`
			Hits []struct {
				ID string `json:"_id"` // 文档ID（即文章UUID）
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	// 将文档ID转换为UUID列表
	// 只返回ID，不返回完整文档，减少网络传输和内存占用
	// 后续可以根据ID从数据库查询完整文章信息
	var ids []uuid.UUID
	for _, h := range result.Hits.Hits {
		if id, err := uuid.Parse(h.ID); err == nil {
			ids = append(ids, id)
		}
		// 如果ID解析失败，跳过该结果（记录日志但不中断流程）
	}

	return ids, result.Hits.Total.Value, nil
}
