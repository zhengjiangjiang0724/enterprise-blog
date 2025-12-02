/**
 * CommentSection 组件 - 评论区域
 * 
 * 功能说明：
 * - 显示文章评论列表（分页）
 * - 支持发表评论（登录用户和匿名用户）
 * - 评论提交后实时更新列表和评论数
 * - 支持分页浏览评论
 * 
 * 技术要点：
 * - 使用 useEffect 监听 articleId 和 page 变化，自动加载评论
 * - 使用回调函数通知父组件更新评论数
 * - 登录用户自动填充用户名和邮箱，匿名用户需要手动输入
 * - 评论提交成功后重置表单并重新加载评论列表
 * 
 * 面试要点：
 * - 为什么使用回调函数？实现组件间通信，解耦组件
 * - 如何实现实时更新？提交成功后重新加载数据，并通过回调通知父组件
 * - 如何处理分页？使用 page 状态，改变时触发 useEffect 重新加载
 */
import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, Comment, PaginatedResponse } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

interface CommentSectionProps {
  articleId: string;              // 文章ID
  /** 评论添加后的回调函数，用于通知父组件更新评论数 */
  onCommentAdded?: () => void;
}

export function CommentSection({ articleId, onCommentAdded }: CommentSectionProps) {
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

  const fetchComments = async () => {
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
  };

  useEffect(() => {
    fetchComments();
  }, [articleId, page, pageSize]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  /**
   * 提交评论处理函数
   * 
   * 流程：
   * 1. 验证评论内容（不能为空）
   * 2. 构建请求载荷（登录用户自动填充信息）
   * 3. 调用API提交评论
   * 4. 成功后：
   *    - 重置表单
   *    - 通知父组件更新评论数（通过回调）
   *    - 重置到第一页并重新加载评论（显示新评论）
   * 
   * 实时更新实现：
   * - 通过 onCommentAdded 回调通知父组件更新评论数
   * - 重置 page 到 1，触发 useEffect 重新加载评论列表
   * - 这样用户可以看到刚提交的评论，无需手动刷新
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault(); // 阻止表单默认提交行为
    
    // 验证评论内容
    if (!content.trim()) {
      showError("评论内容不能为空");
      return;
    }
    
    try {
      // 构建请求载荷
      // 登录用户自动使用用户信息，匿名用户使用表单输入的信息
      const payload = {
        article_id: articleId,
        content,
        author: token && user ? user.username : author || "匿名",
        email: token && user ? user.email : email || undefined
      };
      
      // 提交评论
      const res = await apiClient.post<ApiResponse<Comment>>(
        `/articles/${articleId}/comments`,
        payload
      );
      if (res.data.code !== 200 && res.data.code !== 201) {
        throw new Error(res.data.message || "发表评论失败");
      }
      
      // 成功后处理
      showSuccess("评论已提交");
      setContent(""); // 清空评论内容
      if (!token) {
        // 匿名用户清空表单
        setAuthor("");
        setEmail("");
      }
      
      // 通知父组件更新评论数（实时更新）
      if (onCommentAdded) {
        onCommentAdded();
      }
      
      // 重置到第一页并重新加载评论（显示新评论）
      // fetchComments 会在 page 改变时通过 useEffect 自动触发
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


