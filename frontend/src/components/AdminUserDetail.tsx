import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, UserProfile } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { BackButton } from "./BackButton";
import { useMessage } from "./MessageProvider";

export function AdminUserDetail() {
  const { id } = useParams<{ id: string }>();
  const { token, user } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const { showSuccess, showError } = useMessage();

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

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !profile) return;
    setSaving(true);
    try {
      const payload = {
        role: profile.role,
        status: profile.status
      };
      const res = await apiClient.put<ApiResponse<UserProfile>>(
        `/admin/users/${id}`,
        payload
      );
      if (res.data.code !== 200 || !res.data.data) {
        throw new Error(res.data.message || "保存失败");
      }
      setProfile(res.data.data);
      showSuccess("用户信息已更新");
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "保存失败";
      showError(msg);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="admin-section">
      <div style={{ marginBottom: "16px" }}>
        <BackButton to="/admin/users" label="返回用户列表" />
      </div>
      <h2>用户详情</h2>
      {loading && <p>正在加载用户信息...</p>}
      {error && <p className="error">{error}</p>}
      {profile && (
        <form className="admin-card" onSubmit={handleSave}>
          <p>
            <strong>用户名：</strong> {profile.username}
          </p>
          <p>
            <strong>邮箱：</strong> {profile.email}
          </p>
          <p>
            <strong>角色：</strong>{" "}
            <select
              value={profile.role}
              onChange={(e) =>
                setProfile((prev) =>
                  prev ? { ...prev, role: e.target.value } : prev
                )
              }
            >
              <option value="admin">管理员</option>
              <option value="editor">编辑</option>
              <option value="author">作者</option>
              <option value="reader">读者</option>
            </select>
          </p>
          <p>
            <strong>状态：</strong>{" "}
            <select
              value={profile.status}
              onChange={(e) =>
                setProfile((prev) =>
                  prev ? { ...prev, status: e.target.value } : prev
                )
              }
            >
              <option value="active">启用</option>
              <option value="disabled">禁用</option>
            </select>
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
          <div style={{ marginTop: "12px" }}>
            <Button type="submit" loading={saving}>
              保存变更
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}


