/**
 * API 客户端配置
 * 使用 Axios 创建 HTTP 客户端，配置基础 URL、超时时间
 * 自动在请求头中添加 JWT token（如果存在）
 */
import axios from "axios";

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

/**
 * API 客户端实例
 * - baseURL: API 基础地址，可通过环境变量 VITE_API_BASE_URL 配置
 * - timeout: 请求超时时间 60 秒
 */
export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  // 将超时时间从 10s 提高到 60s，避免后端稍慢时前端过早报超时
  timeout: 60000
});

/**
 * 请求拦截器
 * 自动从 localStorage 读取 token 并添加到请求头中
 */
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});


