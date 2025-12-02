/**
 * ArticleDetail 组件 - 文章详情页
 * 
 * 功能说明：
 * - 显示文章完整内容（标题、封面、正文、统计信息）
 * - 支持点赞、收藏、评论等交互功能
 * - 实时更新点赞数、评论数（无需刷新页面）
 * - 支持编辑和删除（仅作者和管理员）
 * - 集成评论组件，支持发表和查看评论
 * 
 * 技术要点：
 * - 使用 React Hooks（useState、useEffect）管理状态
 * - 使用 React Router 的 useParams 获取路由参数
 * - 使用 useAuth Hook 获取用户认证信息
 * - 使用 useFavorites Hook 管理本地收藏状态
 * - 使用 ReactMarkdown 渲染 Markdown 内容
 * - 实时更新：操作成功后立即更新本地状态，提供即时反馈
 * 
 * 面试要点：
 * - 为什么使用函数式组件？更简洁，性能更好，易于测试
 * - 如何实现实时更新？操作成功后使用 setState 更新本地状态
 * - 如何处理图片加载失败？使用 onError 事件处理，隐藏失败的图片
 */
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { BackButton } from "./BackButton";
import { useMessage } from "./MessageProvider";
import ReactMarkdown from "react-markdown";
import { CommentSection } from "./CommentSection";
import { useFavorites } from "../hooks/useFavorites";

export function ArticleDetail() {
  // 从路由参数获取文章ID
  const { id } = useParams<{ id: string }>();
  
  // 状态管理
  const [article, setArticle] = useState<Article | null>(null); // 文章数据
  const [loading, setLoading] = useState(false);                // 加载状态
  const [error, setError] = useState<string | null>(null);      // 错误信息
  const [liking, setLiking] = useState(false);                  // 点赞中状态（防止重复点击）
  
  // Hooks
  const { token, user } = useAuth();                            // 用户认证信息
  const navigate = useNavigate();                                // 路由导航
  const { showSuccess, showError } = useMessage();             // 消息提示
  const { isFavorited, toggleFavorite } = useFavorites();       // 收藏管理（本地存储）

  // 加载文章数据
  // useEffect 依赖 [id]，当文章ID变化时重新加载
  useEffect(() => {
    if (!id) return; // 如果没有ID，不执行加载
    
    async function fetchArticle() {
      setLoading(true);
      setError(null);
      try {
        // 调用API获取文章详情
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

  // 权限判断：只有管理员或文章作者可以编辑/删除
  const canEditOrDelete =
    !!token && !!user && (user.role === "admin" || user.id === article.author_id);

  /**
   * 删除文章处理函数
   * 
   * 流程：
   * 1. 确认对话框（防止误删）
   * 2. 调用删除API
   * 3. 成功后显示提示并跳转到文章列表
   * 4. 失败时显示错误信息
   */
  const handleDelete = async () => {
    if (!article) return;
    if (!window.confirm("确定要删除这篇文章吗？")) return; // 二次确认，防止误删
    
    try {
      await apiClient.delete(`/articles/${article.id}`);
      showSuccess("文章已删除");
      navigate("/articles"); // 删除成功后跳转到文章列表
    } catch (e: any) {
      const msg =
        e?.response?.data?.message || e.message || "删除失败";
      showError(msg);
    }
  };

  /**
   * 点赞处理函数
   * 
   * 流程：
   * 1. 防止重复点击（liking 状态）
   * 2. 调用点赞API
   * 3. 成功后立即更新本地状态（实时更新点赞数）
   * 4. 显示成功提示
   * 
   * 实时更新实现：
   * - 使用 setState 的函数形式，基于前一个状态更新
   * - 这样可以在不重新请求数据的情况下更新UI
   * - 提供即时的用户反馈，提升用户体验
   */
  const handleLike = async () => {
    if (!article || liking) return; // 防止重复点击
    
    setLiking(true); // 设置点赞中状态，禁用按钮
    try {
      await apiClient.post(`/articles/${article.id}/like`);
      
      // 实时更新点赞数（乐观更新）
      // 使用函数形式的 setState，基于前一个状态更新
      // 这样可以在不重新请求数据的情况下更新UI
      setArticle((prev) =>
        prev ? { ...prev, like_count: prev.like_count + 1 } : prev
      );
      showSuccess("点赞成功");
    } catch (e: any) {
      const msg =
        e?.response?.data?.message || e.message || "点赞失败";
      showError(msg);
    } finally {
      setLiking(false); // 恢复按钮状态
    }
  };

  return (
    <>
      <div style={{ marginBottom: "16px" }}>
        <BackButton to="/" label="返回文章列表" />
      </div>
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
            onClick={() => {
              const wasFavorited = isFavorited(article.id);
              toggleFavorite(article.id);
              showSuccess(wasFavorited ? "已取消收藏" : "已收藏");
            }}
          >
            {isFavorited(article.id) ? "已收藏" : "收藏"}
          </Button>
        </div>
        {/* 封面图片显示 */}
        {/* 处理图片URL：如果是完整URL则直接使用，否则拼接基础URL */}
        {/* 这样支持外部图片链接和内部上传的图片 */}
        {article.cover_image && (
          <img 
            src={article.cover_image.startsWith("http") 
              ? article.cover_image 
              : `${import.meta.env.VITE_API_BASE_URL?.replace("/api/v1", "") || "http://localhost:8080"}${article.cover_image}`} 
            alt={article.title}
            onError={(e) => {
              // 图片加载失败时隐藏图片，避免显示破损图片图标
              // 这是优雅降级的处理方式
              (e.target as HTMLImageElement).style.display = "none";
            }}
          />
        )}
        
        {/* 文章正文内容（Markdown渲染） */}
        <div className="content markdown-content">
          <ReactMarkdown>{article.content}</ReactMarkdown>
        </div>
      </article>
      
      {/* 评论组件 */}
      {/* 通过 onCommentAdded 回调实现实时更新评论数 */}
      <CommentSection 
        articleId={article.id} 
        onCommentAdded={() => {
          // 评论添加后的回调函数
          // 实时更新文章的评论数，无需重新请求文章数据
          // 使用函数形式的 setState，基于前一个状态更新
          setArticle((prev) =>
            prev ? { ...prev, comment_count: prev.comment_count + 1 } : prev
          );
        }}
      />
    </>
  );
}


