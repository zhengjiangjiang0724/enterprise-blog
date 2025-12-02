/**
 * 认证相关的 Hook 和 Context
 * 提供用户登录状态管理、token 存储、用户信息管理等功能
 */
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode
} from "react";

/**
 * 认证用户信息接口
 */
export interface AuthUser {
  id: string;
  username: string;
  email: string;
  role: string;
}

/**
 * 认证上下文值接口
 */
interface AuthContextValue {
  token: string | null;
  user: AuthUser | null;
  login: (token: string, user?: AuthUser | null) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * AuthProvider 认证上下文提供者
 * 管理用户的登录状态，自动同步 token 和用户信息到 localStorage
 * @param children - 子组件
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() =>
    typeof window !== "undefined" ? localStorage.getItem("token") : null
  );
  const [user, setUser] = useState<AuthUser | null>(() => {
    if (typeof window === "undefined") return null;
    const raw = localStorage.getItem("user");
    if (!raw) return null;
    try {
      return JSON.parse(raw) as AuthUser;
    } catch {
      return null;
    }
  });

  useEffect(() => {
    if (token) {
      localStorage.setItem("token", token);
    } else {
      localStorage.removeItem("token");
    }
  }, [token]);

  useEffect(() => {
    if (user) {
      localStorage.setItem("user", JSON.stringify(user));
    } else {
      localStorage.removeItem("user");
    }
  }, [user]);

  /**
   * 登录函数
   * @param newToken - JWT token
   * @param userInfo - 用户信息（可选）
   */
  const login = (newToken: string, userInfo?: AuthUser | null) => {
    setToken(newToken);
    if (userInfo) {
      setUser(userInfo);
    }
  };

  /**
   * 登出函数
   * 清除 token 和用户信息
   */
  const logout = () => {
    setToken(null);
    setUser(null);
  };

  const value = useMemo(
    () => ({
      token,
      user,
      login,
      logout
    }),
    [token, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

/**
 * useAuth Hook
 * 获取认证上下文，提供 token、user、login、logout 等功能
 * @returns 认证上下文值
 * @throws 如果不在 AuthProvider 内使用会抛出错误
 */
export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}


