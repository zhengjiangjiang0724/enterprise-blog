import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { Article, PaginatedResponse } from "../api/types";

export function ArticleList() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") || 1);
  const pageSize = Number(searchParams.get("page_size") || 10);
  const search = searchParams.get("search") || "";

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
        if (res.data.code !== 200) {
          throw new Error(res.data.message || "Failed to load articles");
        }
        setArticles(res.data.data);
        setTotal(res.data.meta?.total || 0);
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

  return (
    <div className="article-list">
      <form className="search-bar" onSubmit={onSearchSubmit}>
        <input
          name="search"
          defaultValue={search}
          placeholder="Search articles..."
        />
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
                  Views: {a.view_count} · Comments: {a.comment_count}
                </p>
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


