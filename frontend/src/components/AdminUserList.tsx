import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { PaginatedResponse, UserProfile } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function AdminUserList() {
  const { token, user } = useAuth();
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") || 1);
  const pageSize = Number(searchParams.get("page_size") || 10);

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!isAdmin) return;
    async function fetchUsers() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<PaginatedResponse<UserProfile>>(
          "/admin/users",
          {
            params: { page, page_size: pageSize }
          }
        );
        if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
          throw new Error(res.data.message || "用户列表加载失败");
        }
        setUsers(res.data.data ?? []);
        setTotal(res.data.meta?.total ?? 0);
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchUsers();
  }, [isAdmin, page, pageSize]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  return (
    <div className="admin-section">
      <h2>用户管理</h2>
      {loading && <p>用户数据加载中...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && (
        <>
          <table className="admin-table">
            <thead>
              <tr>
                <th>用户名</th>
                <th>邮箱</th>
                <th>角色</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id}>
                  <td>{u.username}</td>
                  <td>{u.email}</td>
                  <td>{u.role}</td>
                  <td>{u.status}</td>
                  <td>
                    <Link to={`/admin/users/${u.id}`}>查看</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="pagination">
            <button
              disabled={page <= 1}
              onClick={() =>
                setSearchParams((prev) => {
                  const next = new URLSearchParams(prev);
                  next.set("page", String(page - 1));
                  return next;
                })
              }
              >
                上一页
            </button>
            <span>
                第 {page} / {totalPages} 页
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() =>
                setSearchParams((prev) => {
                  const next = new URLSearchParams(prev);
                  next.set("page", String(page + 1));
                  return next;
                })
              }
              >
                下一页
            </button>
          </div>
        </>
      )}
    </div>
  );
}


