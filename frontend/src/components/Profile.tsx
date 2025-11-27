import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, UserProfile } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { useMessage } from "./MessageProvider";

export function Profile() {
  const { token } = useAuth();
  const { showSuccess, showError } = useMessage();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [form, setForm] = useState({
    username: "",
    email: "",
    avatar: "",
    bio: ""
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (!token) return;
    async function fetchProfile() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<UserProfile>>(
          "/users/profile"
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "个人资料加载失败");
        }
        setProfile(res.data.data);
        setForm({
          username: res.data.data.username || "",
          email: res.data.data.email || "",
          avatar: res.data.data.avatar || "",
          bio: res.data.data.bio || ""
        });
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchProfile();
  }, [token]);

  const handleChange =
    (field: keyof typeof form) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm((prev) => ({
        ...prev,
        [field]: e.target.value
      }));
    };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      const res = await apiClient.put<ApiResponse<UserProfile>>(
        "/users/profile",
        {
          username: form.username,
          email: form.email,
          avatar: form.avatar,
          bio: form.bio
        }
      );
      if (res.data.code !== 200 || !res.data.data) {
        throw new Error(res.data.message || "更新个人资料失败");
      }
      setProfile(res.data.data);
      setSuccess("个人资料已更新");
      showSuccess("个人资料已更新");
    } catch (e: any) {
      const msg =
        e.response?.data?.message || e.message || "发生未知错误";
      setError(msg);
      showError(msg);
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="profile-form">
        <p className="error">请先登录以查看个人资料。</p>
      </div>
    );
  }

  return (
    <div className="profile-form">
      <h2>个人资料</h2>
      {loading && !profile && <p>个人资料加载中...</p>}
      {error && <p className="error">{error}</p>}
      {success && <p className="success">{success}</p>}
      {profile && (
        <form onSubmit={handleSubmit}>
          <label>
            用户名
            <input
              value={form.username}
              onChange={handleChange("username")}
              required
            />
          </label>
          <label>
            邮箱
            <input
              type="email"
              value={form.email}
              onChange={handleChange("email")}
              required
            />
          </label>
          <label>
            头像地址
            <input value={form.avatar} onChange={handleChange("avatar")} />
          </label>
          <label>
            个性签名
            <textarea
              rows={4}
              value={form.bio}
              onChange={handleChange("bio")}
            />
          </label>
          <button type="submit" className="button primary" disabled={loading}>
            {loading ? "保存中..." : "更新资料"}
          </button>
        </form>
      )}
    </div>
  );
}

