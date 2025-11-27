import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, Comment, PaginatedResponse } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

interface CommentSectionProps {
  articleId: string;
}

export function CommentSection({ articleId }: CommentSectionProps) {
  const { token, user } = useAuth();
  const { showError, showSuccess } = useMessage();

  const [comments, setComments] = useState<Comment[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [content, setContent] = useState("");
  const [author, setAuthor] = useState("");
  const [email, setEmail] = useState("");

  useEffect(() => {
    async function fetchComments() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<PaginatedResponse<Comment>>(
          `/articles/${articleId}/comments`,
          {
            params: { page, page_size: pageSize }
          }
        );
        if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
          throw new Error(res.data.message || "评论加载失败");
        }
        setComments(res.data.data ?? []);
        setTotal(res.data.meta?.total ?? 0);
      } catch (e: any) {
        const msg = e.message || "评论加载失败";
        setError(msg);
      } finally {
        setLoading(false);
      }
    }
    fetchComments();
  }, [articleId, page, pageSize]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      showError("评论内容不能为空");
      return;
    }
    try {
      const payload = {
        article_id: articleId,
        content,
        author: token && user ? user.username : author || "匿名",
        email: token && user ? user.email : email || undefined
      };
      const res = await apiClient.post<ApiResponse<Comment>>(
        `/articles/${articleId}/comments`,
        payload
      );
      if (res.data.code !== 200 && res.data.code !== 201) {
        throw new Error(res.data.message || "发表评论失败");
      }
      showSuccess("评论已提交");
      setContent("");
      if (!token) {
        setAuthor("");
        setEmail("");
      }
      // 重新加载第一页评论
      setPage(1);
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "发表评论失败";
      showError(msg);
    }
  };

  return (
    <section className="comments-section">
      <h2>评论</h2>
      {loading && <p>评论加载中...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && comments.length === 0 && (
        <p className="meta">还没有评论，快来抢沙发吧。</p>
      )}
      {comments.length > 0 && (
        <ul className="comment-list">
          {comments.map((c) => (
            <li key={c.id} className="comment-item">
              <div className="comment-author">
                {c.author || "匿名"}{" "}
                <span className="comment-date">
                  {new Date(c.created_at).toLocaleString()}
                </span>
              </div>
              <div className="comment-content">{c.content}</div>
            </li>
          ))}
        </ul>
      )}

      {totalPages > 1 && (
        <div className="pagination">
          <button
            disabled={page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            上一页
          </button>
          <span>
            第 {page} / {totalPages} 页
          </span>
          <button
            disabled={page >= totalPages}
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          >
            下一页
          </button>
        </div>
      )}

      <div className="comment-form">
        <h3>发表评论</h3>
        <form onSubmit={handleSubmit}>
          {!token && (
            <>
              <label>
                昵称
                <input
                  value={author}
                  onChange={(e) => setAuthor(e.target.value)}
                  placeholder="选填，默认匿名"
                />
              </label>
              <label>
                邮箱
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="选填，用于通知"
                />
              </label>
            </>
          )}
          <label>
            内容
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={4}
              required
            />
          </label>
          <Button type="submit" variant="primary">
            提交评论
          </Button>
        </form>
      </div>
    </section>
  );
}


