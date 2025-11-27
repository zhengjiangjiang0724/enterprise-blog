import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function ArticleDetail() {
  const { id } = useParams<{ id: string }>();
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { token } = useAuth();

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

  if (loading) return <p>Loading article...</p>;
  if (error) return <p className="error">{error}</p>;
  if (!article) return <p>Article not found.</p>;

  return (
    <article className="article-detail">
      <h1>{article.title}</h1>
      {token && (
        <div className="article-actions">
          <Link to={`/articles/${article.id}/edit`} className="button secondary">
            Edit Article
          </Link>
        </div>
      )}
      <p className="meta">
        {article.category && <span>Category: {article.category.name} · </span>}
        Views: {article.view_count} · Comments: {article.comment_count}
      </p>
      {article.tags && article.tags.length > 0 && (
        <p className="meta">
          Tags:{" "}
          {article.tags.map((t, idx) => (
            <span key={t.id}>
              {t.name}
              {idx < article.tags!.length - 1 ? ", " : ""}
            </span>
          ))}
        </p>
      )}
      {article.cover_image && (
        <img src={article.cover_image} alt={article.title} />
      )}
      <div className="content">{article.content}</div>
    </article>
  );
}


