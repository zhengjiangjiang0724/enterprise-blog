import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

export function AdminArticleDetail() {
  const { id } = useParams<{ id: string }>();
  const { token, user } = useAuth();
  const { showSuccess, showError } = useMessage();
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState(false);

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!token || !isAdmin || !id) return;
    async function fetchArticle() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<Article>>(
          `/admin/articles/${id}`
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "文章信息加载失败");
        }
        setArticle(res.data.data);
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchArticle();
  }, [token, isAdmin, id]);

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  const handleStatusChange = async (status: string) => {
    if (!id) return;
    setUpdating(true);
    try {
      const res = await apiClient.put<ApiResponse<Article>>(
        `/admin/articles/${id}/status`,
        { status }
      );
      if (res.data.code !== 200 || !res.data.data) {
        throw new Error(res.data.message || "更新状态失败");
      }
      setArticle(res.data.data);
      showSuccess("文章状态已更新");
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "更新状态失败";
      showError(msg);
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = async () => {
    if (!id) return;
    if (!window.confirm("确定要删除这篇文章吗？此操作不可恢复。")) return;
    setUpdating(true);
    try {
      const res = await apiClient.delete<ApiResponse<null>>(
        `/admin/articles/${id}`
      );
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "删除失败");
      }
      showSuccess("文章已删除");
      setArticle(null);
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "删除失败";
      showError(msg);
    } finally {
      setUpdating(false);
    }
  };

  return (
    <div className="admin-section">
      <h2>文章详情</h2>
      {loading && <p>正在加载文章信息...</p>}
      {error && <p className="error">{error}</p>}
      {article && (
        <div className="admin-card">
          <p>
            <strong>标题：</strong> {article.title}
          </p>
          <p>
            <strong>状态：</strong> {article.status}
          </p>
          <p>
            <strong>阅读 / 点赞 / 评论：</strong>{" "}
            {article.view_count} / {article.like_count} /{" "}
            {article.comment_count}
          </p>
          <p>
            <strong>创建时间：</strong>{" "}
            {new Date(article.created_at).toLocaleString()}
          </p>
          <p>
            <strong>更新时间：</strong>{" "}
            {new Date(article.updated_at).toLocaleString()}
          </p>
          <p>
            <strong>摘要：</strong> {article.excerpt || "（暂无）"}
          </p>
          <p>
            <strong>正文（前 200 字）：</strong>{" "}
            {article.content.slice(0, 200)}...
          </p>
          <div style={{ marginTop: "12px", display: "flex", gap: "8px" }}>
            <Button
              type="button"
              variant="secondary"
              disabled={updating}
              onClick={() => handleStatusChange("draft")}
            >
              设为草稿
            </Button>
            <Button
              type="button"
              variant="primary"
              disabled={updating}
              onClick={() => handleStatusChange("published")}
            >
              发布
            </Button>
            <Button
              type="button"
              variant="secondary"
              disabled={updating}
              onClick={() => handleStatusChange("archived")}
            >
              归档
            </Button>
            <Button
              type="button"
              variant="danger"
              disabled={updating}
              onClick={handleDelete}
            >
              删除文章
            </Button>
          </div>
        </div>
      )}
      {!loading && !error && !article && (
        <p className="meta">文章已删除或不存在。</p>
      )}
    </div>
  );
}


