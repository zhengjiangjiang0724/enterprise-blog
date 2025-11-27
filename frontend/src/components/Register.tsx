import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse } from "../api/types";
import { useAuth, AuthUser } from "../hooks/useAuth";

interface RegisterResponse {
  id: string;
  username: string;
  email: string;
}

export function Register() {
  const navigate = useNavigate();
  const { login } = useAuth();

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
        throw new Error(res.data.message || "Registration failed");
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

      navigate("/profile");
    } catch (e: any) {
      setError(e.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-form">
      <h2>Register</h2>
      <form onSubmit={handleSubmit}>
        <label>
          Username
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </label>
        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={loading}>
          {loading ? "Registering..." : "Register"}
        </button>
      </form>
      <p className="auth-switch">
        Already have an account? <Link to="/login">Login</Link>
      </p>
    </div>
  );
}

