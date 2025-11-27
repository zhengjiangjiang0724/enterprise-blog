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
  const categoryId = searchParams.get("category_id") || "";
  const tagId = searchParams.get("tag_id") || "";

  // 加载分类和标签
  useEffect(() => {
    async function fetchCategoriesAndTags() {
      try {
        const [catRes, tagRes] = await Promise.all([
          apiClient.get<PaginatedResponse<Category>>("/categories"),
          apiClient.get<PaginatedResponse<Tag>>("/tags")
        ]);
        if (Array.isArray(catRes.data.data)) {
          setCategories(catRes.data.data);
        }
        if (Array.isArray(tagRes.data.data)) {
          setTags(tagRes.data.data);
        }
      } catch {
        // 分类/标签加载失败不影响文章列表，静默处理或在需要时展示提示
      }
    }
    fetchCategoriesAndTags();
  }, []);

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
              search,
              category_id: categoryId || undefined,
              tag_id: tagId || undefined
            }
          }
        );
        if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
          throw new Error(res.data.message || "Failed to load articles");
        }
        setArticles(res.data.data ?? []);
        setTotal(res.data.meta?.total ?? 0);
      } catch (e: any) {
        setError(e.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }

    fetchArticles();
  }, [page, pageSize, search]);

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

  const handleCategoryChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set("category_id", value);
    } else {
      next.delete("category_id");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  const handleTagChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set("tag_id", value);
    } else {
      next.delete("tag_id");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  return (
    <div className="article-list">
      <form className="search-bar" onSubmit={onSearchSubmit}>
        <input
          name="search"
          defaultValue={search}
          placeholder="Search articles..."
        />
        <select
          value={categoryId}
          onChange={handleCategoryChange}
          className="filter-select"
        >
          <option value="">All Categories</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
        <select
          value={tagId}
          onChange={handleTagChange}
          className="filter-select"
        >
          <option value="">All Tags</option>
          {tags.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
        <button type="submit">Search</button>
      </form>

      {loading && <p>Loading articles...</p>}
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
                  {a.category && <span>Category: {a.category.name} · </span>}
                  Views: {a.view_count} · Comments: {a.comment_count}
                </p>
                {a.tags && a.tags.length > 0 && (
                  <p className="meta">
                    Tags:{" "}
                    {a.tags.map((t, idx) => (
                      <span key={t.id}>
                        {t.name}
                        {idx < a.tags!.length - 1 ? ", " : ""}
                      </span>
                    ))}
                  </p>
                )}
                <p>{a.excerpt}</p>
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
              Prev
            </button>
            <span>
              Page {page} / {totalPages}
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
              Next
            </button>
          </div>
        </>
      )}
    </div>
  );
}


