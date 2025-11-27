import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse } from "../api/types";
import { useAuth, AuthUser } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

interface RegisterResponse {
  id: string;
  username: string;
  email: string;
}

export function Register() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const { showSuccess, showError } = useMessage();

  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.post<ApiResponse<RegisterResponse>>(
        "/auth/register",
        {
          username,
          email,
          password
        }
      );
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "注册失败");
      }

      // 注册后自动登录
      const loginRes =
        await apiClient.post<ApiResponse<{ token: string; user?: AuthUser }>>(
          "/auth/login",
          {
            email,
            password
          }
        );
      if (loginRes.data.code === 200 && loginRes.data.data.token) {
        const loginUser: AuthUser =
          loginRes.data.data.user ?? {
            id: res.data.data.id,
            username: res.data.data.username,
            email: res.data.data.email,
            role: "reader"
          };
        try {
          localStorage.setItem("token", loginRes.data.data.token);
          localStorage.setItem("user", JSON.stringify(loginUser));
        } catch {
          // 忽略存储错误
        }
        login(loginRes.data.data.token, loginUser);
      }

      showSuccess("注册并登录成功");
      navigate("/profile");
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
      <h2>注册账号</h2>
      <form onSubmit={handleSubmit}>
        <label>
          用户名
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </label>
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
          注册
        </Button>
      </form>
      <p className="auth-switch">
        已有账号？<Link to="/login">前往登录</Link>
      </p>
    </div>
  );
}

