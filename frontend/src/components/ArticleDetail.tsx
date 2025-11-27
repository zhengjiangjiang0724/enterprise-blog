import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article } from "../api/types";

export function ArticleDetail() {
  const { id } = useParams<{ id: string }>();
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
      <p className="meta">
        Views: {article.view_count} · Comments: {article.comment_count}
      </p>
      {article.cover_image && (
        <img src={article.cover_image} alt={article.title} />
      )}
      <div className="content">{article.content}</div>
    </article>
  );
}


