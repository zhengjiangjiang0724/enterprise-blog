import { useState, useEffect } from "react";
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

type LoginMode = "email" | "phone";

export function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const { showSuccess, showError } = useMessage();

  const [mode, setMode] = useState<LoginMode>("email");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [sendingCode, setSendingCode] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [error, setError] = useState<string | null>(null);

  // 倒计时
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [countdown]);

  const handleSendCode = async () => {
    if (!phone.trim()) {
      showError("请输入手机号");
      return;
    }
    setSendingCode(true);
    setError(null);
    try {
      const res = await apiClient.post<ApiResponse<{ message: string }>>(
        "/auth/send-sms-code",
        { phone }
      );
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "发送验证码失败");
      }
      showSuccess("验证码已发送");
      setCountdown(60); // 60秒倒计时
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "发送验证码失败";
      setError(msg);
      showError(msg);
    } finally {
      setSendingCode(false);
    }
  };

  const handleEmailLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.post<ApiResponse<LoginResponse>>(
        "/auth/login",
        { email, password }
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
      try {
        localStorage.setItem("token", res.data.data.token);
        if (authUser) {
          localStorage.setItem("user", JSON.stringify(authUser));
        }
      } catch {
        // 忽略存储错误
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

  const handlePhoneLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.post<ApiResponse<LoginResponse>>(
        "/auth/login-phone",
        { phone, code }
      );
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "登录失败");
      }
      const authUser: AuthUser | undefined = res.data.data.user
        ? {
            id: res.data.data.user.id,
            username: res.data.data.user.username,
            email: res.data.data.user.email || "",
            role: (res.data.data.user as any).role || "reader"
          }
        : undefined;
      try {
        localStorage.setItem("token", res.data.data.token);
        if (authUser) {
          localStorage.setItem("user", JSON.stringify(authUser));
        }
      } catch {
        // 忽略存储错误
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
      <div style={{ display: "flex", gap: "8px", marginBottom: "16px" }}>
        <button
          type="button"
          onClick={() => setMode("email")}
          style={{
            padding: "8px 16px",
            border: "1px solid #ddd",
            background: mode === "email" ? "#007bff" : "white",
            color: mode === "email" ? "white" : "#333",
            cursor: "pointer",
            borderRadius: "4px"
          }}
        >
          邮箱登录
        </button>
        <button
          type="button"
          onClick={() => setMode("phone")}
          style={{
            padding: "8px 16px",
            border: "1px solid #ddd",
            background: mode === "phone" ? "#007bff" : "white",
            color: mode === "phone" ? "white" : "#333",
            cursor: "pointer",
            borderRadius: "4px"
          }}
        >
          手机登录
        </button>
      </div>

      {mode === "email" ? (
        <form onSubmit={handleEmailLogin}>
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
      ) : (
        <form onSubmit={handlePhoneLogin}>
          <label>
            手机号
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="请输入手机号"
              required
            />
          </label>
          <label>
            验证码
            <div style={{ display: "flex", gap: "8px" }}>
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="请输入6位验证码"
                maxLength={6}
                required
                style={{ flex: 1 }}
              />
              <Button
                type="button"
                variant="secondary"
                onClick={handleSendCode}
                disabled={sendingCode || countdown > 0}
                loading={sendingCode}
              >
                {countdown > 0 ? `${countdown}秒` : "发送验证码"}
              </Button>
            </div>
          </label>
          {error && <p className="error">{error}</p>}
          <Button type="submit" loading={loading}>
            登录
          </Button>
        </form>
      )}

      <p className="auth-switch">
        还没有账号？<Link to="/register">立即注册</Link>
      </p>
    </div>
  );
}


