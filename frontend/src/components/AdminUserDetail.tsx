import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, UserProfile } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function AdminUserDetail() {
  const { id } = useParams<{ id: string }>();
  const { token, user } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!token || !isAdmin || !id) return;
    async function fetchUser() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<UserProfile>>(
          `/admin/users/${id}`
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "用户信息加载失败");
        }
        setProfile(res.data.data);
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchUser();
  }, [token, isAdmin, id]);

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  return (
    <div className="admin-section">
      <h2>用户详情</h2>
      {loading && <p>正在加载用户信息...</p>}
      {error && <p className="error">{error}</p>}
      {profile && (
        <div className="admin-card">
          <p>
            <strong>用户名：</strong> {profile.username}
          </p>
          <p>
            <strong>邮箱：</strong> {profile.email}
          </p>
          <p>
            <strong>角色：</strong> {profile.role}
          </p>
          <p>
            <strong>状态：</strong> {profile.status}
          </p>
          <p>
            <strong>签名：</strong> {profile.bio || "-"}
          </p>
          <p>
            <strong>创建时间：</strong>{" "}
            {new Date(profile.created_at).toLocaleString()}
          </p>
          <p>
            <strong>更新时间：</strong>{" "}
            {new Date(profile.updated_at).toLocaleString()}
          </p>
        </div>
      )}
    </div>
  );
}


