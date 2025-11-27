import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse } from "../api/types";
import { useAuth, AuthUser } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

interface LoginResponse {
  token: string;
  user: {
    id: string;
    username: string;
    email: string;
    role: string;
  };
}

export function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const { showSuccess, showError } = useMessage();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.post<ApiResponse<LoginResponse>>(
        "/auth/login",
        {
          email,
          password
        }
      );
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "登录失败");
      }
      const authUser: AuthUser | undefined = res.data.data.user
        ? {
            id: res.data.data.user.id,
            username: res.data.data.user.username,
            email: res.data.data.user.email,
            role: (res.data.data.user as any).role || "reader"
          }
        : undefined;
      // 显式写入 localStorage，确保页面刷新后仍能读取到
      try {
        localStorage.setItem("token", res.data.data.token);
        if (authUser) {
          localStorage.setItem("user", JSON.stringify(authUser));
        }
      } catch {
        // 忽略存储错误（如隐私模式）
      }
      login(res.data.data.token, authUser);
      showSuccess("登录成功");
      navigate("/");
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "发生未知错误";
      setError(msg);
      showError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-form">
      <h2>账号登录</h2>
      <form onSubmit={handleSubmit}>
        <label>
          邮箱
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          密码
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        {error && <p className="error">{error}</p>}
        <Button type="submit" loading={loading}>
          登录
        </Button>
      </form>
      <p className="auth-switch">
        还没有账号？<Link to="/register">立即注册</Link>
      </p>
    </div>
  );
}


