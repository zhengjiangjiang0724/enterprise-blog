import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, Category } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

export function AdminCategoryList() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useMessage();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newDesc, setNewDesc] = useState("");

  const isAdmin = user?.role === "admin";

  const loadCategories = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.get<ApiResponse<Category[]>>(
        "/admin/categories"
      );
      if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
        throw new Error(res.data.message || "分类列表加载失败");
      }
      setCategories(res.data.data ?? []);
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "分类列表加载失败";
      setError(msg);
      showError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token || !isAdmin) return;
    loadCategories();
  }, [token, isAdmin]);

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) {
      showError("分类名称不能为空");
      return;
    }
    setCreating(true);
    try {
      const res = await apiClient.post<ApiResponse<Category>>(
        "/admin/categories",
        {
          name: newName,
          description: newDesc || undefined
        }
      );
      if (res.data.code !== 200 && res.data.code !== 201) {
        throw new Error(res.data.message || "创建分类失败");
      }
      showSuccess("分类已创建");
      setNewName("");
      setNewDesc("");
      loadCategories();
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "创建分类失败";
      showError(msg);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("确定要删除该分类吗？删除后关联文章的分类会被置空。")) {
      return;
    }
    try {
      await apiClient.delete<ApiResponse<null>>(`/admin/categories/${id}`);
      showSuccess("分类已删除");
      loadCategories();
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "删除分类失败";
      showError(msg);
    }
  };

  return (
    <div className="admin-section">
      <h2>分类管理</h2>
      <form className="admin-inline-form" onSubmit={handleCreate}>
        <input
          placeholder="分类名称"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
        />
        <input
          placeholder="描述（可选）"
          value={newDesc}
          onChange={(e) => setNewDesc(e.target.value)}
        />
        <Button type="submit" loading={creating}>
          新增分类
        </Button>
      </form>
      {loading && <p>分类加载中...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && (
        <table className="admin-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>Slug</th>
              <th>描述</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {categories.map((c) => (
              <tr key={c.id}>
                <td>{c.name}</td>
                <td>{c.slug}</td>
                <td>{c.description || "-"}</td>
                <td>
                  <Button
                    type="button"
                    variant="danger"
                    size="small"
                    onClick={() => handleDelete(c.id)}
                  >
                    删除
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}


