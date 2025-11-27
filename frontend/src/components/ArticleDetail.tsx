import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";
import ReactMarkdown from "react-markdown";
import { CommentSection } from "./CommentSection";
import { useFavorites } from "../hooks/useFavorites";

export function ArticleDetail() {
  const { id } = useParams<{ id: string }>();
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { token, user } = useAuth();
  const navigate = useNavigate();
  const { showSuccess, showError } = useMessage();
  const [liking, setLiking] = useState(false);
  const { isFavorited, toggleFavorite } = useFavorites();

  useEffect(() => {
    if (!id) return;
    async function fetchArticle() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<Article>>(`/articles/${id}`);
        if (res.data.code !== 200) {
          throw new Error(res.data.message || "Failed to load article");
        }
        setArticle(res.data.data);
      } catch (e: any) {
        setError(e.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }

    fetchArticle();
  }, [id]);

  if (loading) return <p>文章加载中...</p>;
  if (error) return <p className="error">{error}</p>;
  if (!article) return <p>未找到文章。</p>;

  const canEditOrDelete =
    !!token && !!user && (user.role === "admin" || user.id === article.author_id);

  const handleDelete = async () => {
    if (!article) return;
    if (!window.confirm("确定要删除这篇文章吗？")) return;
    try {
      await apiClient.delete(`/articles/${article.id}`);
      showSuccess("文章已删除");
      navigate("/articles");
    } catch (e: any) {
      const msg =
        e?.response?.data?.message || e.message || "删除失败";
      showError(msg);
    }
  };

  const handleLike = async () => {
    if (!article || liking) return;
    setLiking(true);
    try {
      await apiClient.post(`/articles/${article.id}/like`);
      setArticle((prev) =>
        prev ? { ...prev, like_count: prev.like_count + 1 } : prev
      );
    } catch (e: any) {
      const msg =
        e?.response?.data?.message || e.message || "点赞失败";
      showError(msg);
    } finally {
      setLiking(false);
    }
  };

  return (
    <>
      <article className="article-detail">
        <h1>{article.title}</h1>
        {canEditOrDelete && (
          <div className="article-actions">
            <Link
              to={`/articles/${article.id}/edit`}
              className="button secondary"
            >
              编辑文章
            </Link>
            <Button
              type="button"
              variant="danger"
              onClick={handleDelete}
              style={{ marginLeft: "8px" }}
            >
              删除文章
            </Button>
          </div>
        )}
        <p className="meta">
          阅读：{article.view_count} · 点赞：{article.like_count} · 评论：
          {article.comment_count}
        </p>
        <div style={{ marginBottom: "8px", display: "flex", gap: "8px" }}>
          <Button
            type="button"
            variant="secondary"
            onClick={handleLike}
            disabled={liking}
          >
            {liking ? "点赞中..." : "点赞"}
          </Button>
          <Button
            type="button"
            variant={isFavorited(article.id) ? "primary" : "ghost"}
            onClick={() => toggleFavorite(article.id)}
          >
            {isFavorited(article.id) ? "已收藏" : "收藏"}
          </Button>
        </div>
        {article.cover_image && (
          <img src={article.cover_image} alt={article.title} />
        )}
        <div className="content markdown-content">
          <ReactMarkdown>{article.content}</ReactMarkdown>
        </div>
      </article>
      <CommentSection articleId={article.id} />
    </>
  );
}


