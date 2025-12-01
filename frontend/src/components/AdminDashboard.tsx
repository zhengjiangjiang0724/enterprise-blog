import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, AdminDashboardStats } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function AdminDashboard() {
  const { token, user } = useAuth();
  const [stats, setStats] = useState<AdminDashboardStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!token || !isAdmin) return;
    async function fetchStats() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<AdminDashboardStats>>(
          "/admin/dashboard"
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "仪表盘数据加载失败");
        }
        setStats(res.data.data);
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchStats();
  }, [token, isAdmin]);

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  return (
    <div className="admin-section">
      <h2>后台仪表盘</h2>
      {loading && <p>仪表盘数据加载中...</p>}
      {error && <p className="error">{error}</p>}
      {stats && !loading && !error && (
        <div className="dashboard-grid">
          <div className="dashboard-card">
            <h3>用户总数</h3>
            <p className="dashboard-number">{stats.total_users}</p>
          </div>
          <div className="dashboard-card">
            <h3>文章总数</h3>
            <p className="dashboard-number">{stats.total_articles}</p>
          </div>
          <div className="dashboard-card">
            <h3>已发布文章</h3>
            <p className="dashboard-number">{stats.published_articles}</p>
          </div>
          <div className="dashboard-card">
            <h3>草稿 / 归档</h3>
            <p className="dashboard-number">
              {stats.draft_articles} / {stats.archived_articles}
            </p>
          </div>
          <div className="dashboard-card">
            <h3>总阅读 / 点赞</h3>
            <p className="dashboard-number">
              {stats.total_article_views} / {stats.total_article_likes}
            </p>
          </div>
          <div className="dashboard-card">
            <h3>评论总数</h3>
            <p className="dashboard-number">{stats.total_comments}</p>
          </div>
          <div className="dashboard-card">
            <h3>今日新发布文章</h3>
            <p className="dashboard-number">{stats.today_published_count}</p>
          </div>
        </div>
      )}
    </div>
  );
}


