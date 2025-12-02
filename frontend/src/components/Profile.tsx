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
  const [pwdLoading, setPwdLoading] = useState(false);
  const [pwdError, setPwdError] = useState<string | null>(null);
  const [pwdSuccess, setPwdSuccess] = useState<string | null>(null);
  const [pwdForm, setPwdForm] = useState({
    old_password: "",
    new_password: "",
    confirm_password: ""
  });

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

  const handlePwdChange =
    (field: keyof typeof pwdForm) =>
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setPwdForm((prev) => ({
        ...prev,
        [field]: e.target.value
      }));
    };

  const handlePwdSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setPwdError(null);
    setPwdSuccess(null);
    if (!pwdForm.old_password || !pwdForm.new_password) {
      setPwdError("请填写完整的旧密码和新密码。");
      return;
    }
    if (pwdForm.new_password !== pwdForm.confirm_password) {
      setPwdError("两次输入的新密码不一致。");
      return;
    }
    setPwdLoading(true);
    try {
      const res = await apiClient.put<ApiResponse<null>>("/users/password", {
        old_password: pwdForm.old_password,
        new_password: pwdForm.new_password
      });
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "修改密码失败");
      }
      setPwdSuccess("密码已修改，请使用新密码登录。");
      showSuccess("密码已修改");
      setPwdForm({
        old_password: "",
        new_password: "",
        confirm_password: ""
      });
    } catch (e: any) {
      const msg =
        e.response?.data?.message || e.message || "修改密码失败";
      setPwdError(msg);
      showError(msg);
    } finally {
      setPwdLoading(false);
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
        <>
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

          <hr style={{ margin: "24px 0" }} />

          <h3>修改密码</h3>
          {pwdError && <p className="error">{pwdError}</p>}
          {pwdSuccess && <p className="success">{pwdSuccess}</p>}
          <form onSubmit={handlePwdSubmit}>
            <label>
              旧密码
              <input
                type="password"
                value={pwdForm.old_password}
                onChange={handlePwdChange("old_password")}
                required
              />
            </label>
            <label>
              新密码
              <input
                type="password"
                value={pwdForm.new_password}
                onChange={handlePwdChange("new_password")}
                required
              />
            </label>
            <label>
              确认新密码
              <input
                type="password"
                value={pwdForm.confirm_password}
                onChange={handlePwdChange("confirm_password")}
                required
              />
            </label>
            <button
              type="submit"
              className="button secondary"
              disabled={pwdLoading}
            >
              {pwdLoading ? "修改中..." : "修改密码"}
            </button>
          </form>
        </>
      )}
    </div>
  );
}

