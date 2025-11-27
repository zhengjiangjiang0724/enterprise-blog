import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { Article, PaginatedResponse } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function AdminArticleList() {
  const { token, user } = useAuth();
  const [articles, setArticles] = useState<Article[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchParams, setSearchParams] = useSearchParams();

  const page = Number(searchParams.get("page") || 1);
  const pageSize = Number(searchParams.get("page_size") || 10);
  const search = searchParams.get("search") || "";
  const status = searchParams.get("status") || "";

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!isAdmin) return;
    async function fetchArticles() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<PaginatedResponse<Article>>(
          "/admin/articles",
          {
            params: {
              page,
              page_size: pageSize,
              search,
              status
            }
          }
        );
        if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
          throw new Error(res.data.message || "文章列表加载失败");
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
  }, [isAdmin, page, pageSize, search, status]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  const handleFilterSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const formData = new FormData(form);
    const next = new URLSearchParams(searchParams);
    const s = String(formData.get("search") || "");
    const st = String(formData.get("status") || "");
    if (s) {
      next.set("search", s);
    } else {
      next.delete("search");
    }
    if (st) {
      next.set("status", st);
    } else {
      next.delete("status");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  return (
    <div className="admin-section">
      <h2>文章管理</h2>
      <form className="search-bar" onSubmit={handleFilterSubmit}>
        <input
          name="search"
          defaultValue={search}
          placeholder="搜索标题或内容关键字..."
        />
        <select
          name="status"
          defaultValue={status}
          className="filter-select"
        >
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="published">已发布</option>
          <option value="archived">已归档</option>
        </select>
        <button type="submit">筛选</button>
      </form>

      {loading && <p>文章数据加载中...</p>}
      {error && <p className="error">{error}</p>}

      {!loading && !error && (
        <>
          <table className="admin-table">
            <thead>
              <tr>
                <th>标题</th>
                <th>状态</th>
                <th>阅读</th>
                <th>点赞</th>
                <th>评论</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {articles.map((a) => (
                <tr key={a.id}>
                  <td>{a.title}</td>
                  <td>{a.status}</td>
                  <td>{a.view_count}</td>
                  <td>{a.like_count}</td>
                  <td>{a.comment_count}</td>
                  <td>{new Date(a.created_at).toLocaleString()}</td>
                  <td>
                    <Link to={`/admin/articles/${a.id}`}>查看</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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


