import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type {
  Article,
  Category,
  PaginatedResponse,
  Tag
} from "../api/types";

export function ArticleList() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") || 1);
  const pageSize = Number(searchParams.get("page_size") || 10);
  const search = searchParams.get("search") || "";
  const categoryId = "";
  const tagId = "";

  // 分类和标签过滤功能已简化，这里不再请求分类/标签列表

  useEffect(() => {
    async function fetchArticles() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<PaginatedResponse<Article>>(
          "/articles",
          {
            params: {
              page,
              page_size: pageSize,
              search
            }
          }
        );
        if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
          throw new Error(res.data.message || "文章加载失败");
        }
        setArticles(res.data.data ?? []);
        setTotal(res.data.meta?.total ?? 0);
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }

    fetchArticles();
  }, [page, pageSize, search, categoryId, tagId]);

  const onSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const value = String(formData.get("search") || "");
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set("search", value);
    } else {
      next.delete("search");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  // 分类和标签筛选已移除

  return (
    <div className="article-list">
      <form className="search-bar" onSubmit={onSearchSubmit}>
        <input
          name="search"
          defaultValue={search}
          placeholder="搜索文章..."
        />
        {/* 分类和标签下拉已简化，仅保留搜索 */}
        <button type="submit">搜索</button>
      </form>

      {loading && <p>正在加载文章...</p>}
      {error && <p className="error">{error}</p>}

      {!loading && !error && (
        <>
          <ul>
            {articles.map((a) => (
              <li key={a.id} className="article-item">
                <h2>
                  <Link to={`/articles/${a.id}`}>{a.title}</Link>
                </h2>
                <p className="meta">
                  阅读：{a.view_count} · 点赞：{a.like_count} · 评论：
                  {a.comment_count}
                </p>
                <p>{a.excerpt || "（暂无摘要）"}</p>
              </li>
            ))}
          </ul>

          <div className="pagination">
            <button
              disabled={page <= 1}
              onClick={() =>
                setSearchParams((prev) => {
                  const next = new URLSearchParams(prev);
                  next.set("page", String(page - 1));
                  return next;
                })
              }
              >
                上一页
            </button>
            <span>
                第 {page} / {totalPages} 页
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() =>
                setSearchParams((prev) => {
                  const next = new URLSearchParams(prev);
                  next.set("page", String(page + 1));
                  return next;
                })
              }
              >
                下一页
            </button>
          </div>
        </>
      )}
    </div>
  );
}


