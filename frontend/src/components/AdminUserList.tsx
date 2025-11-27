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
          throw new Error(res.data.message || "Failed to load users");
        }
        setUsers(res.data.data ?? []);
        setTotal(res.data.meta?.total ?? 0);
      } catch (e: any) {
        setError(e.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchUsers();
  }, [isAdmin, page, pageSize]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (!token) {
    return <p className="error">Please login as admin.</p>;
  }
  if (!isAdmin) {
    return <p className="error">Access denied. Admin only.</p>;
  }

  return (
    <div className="admin-section">
      <h2>User Management</h2>
      {loading && <p>Loading users...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && (
        <>
          <table className="admin-table">
            <thead>
              <tr>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th>Status</th>
                <th>Actions</th>
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
                    <Link to={`/admin/users/${u.id}`}>View</Link>
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
              Prev
            </button>
            <span>
              Page {page} / {totalPages}
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
              Next
            </button>
          </div>
        </>
      )}
    </div>
  );
}


