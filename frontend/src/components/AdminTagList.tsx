import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, Tag } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

export function AdminTagList() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useMessage();
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newColor, setNewColor] = useState("");

  const isAdmin = user?.role === "admin";

  const loadTags = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.get<ApiResponse<Tag[]>>("/admin/tags");
      if (res.data.code !== 200 || !Array.isArray(res.data.data)) {
        throw new Error(res.data.message || "标签列表加载失败");
      }
      setTags(res.data.data ?? []);
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "标签列表加载失败";
      setError(msg);
      showError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token || !isAdmin) return;
    loadTags();
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
      showError("标签名称不能为空");
      return;
    }
    setCreating(true);
    try {
      const res = await apiClient.post<ApiResponse<Tag>>("/admin/tags", {
        name: newName,
        color: newColor || undefined
      });
      if (res.data.code !== 200 && res.data.code !== 201) {
        throw new Error(res.data.message || "创建标签失败");
      }
      showSuccess("标签已创建");
      setNewName("");
      setNewColor("");
      loadTags();
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "创建标签失败";
      showError(msg);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("确定要删除该标签吗？删除后关联文章的该标签会被移除。")) {
      return;
    }
    try {
      await apiClient.delete<ApiResponse<null>>(`/admin/tags/${id}`);
      showSuccess("标签已删除");
      loadTags();
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "删除标签失败";
      showError(msg);
    }
  };

  return (
    <div className="admin-section">
      <h2>标签管理</h2>
      <form className="admin-inline-form" onSubmit={handleCreate}>
        <input
          placeholder="标签名称"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
        />
        <input
          placeholder="颜色（可选，例如 #1890ff）"
          value={newColor}
          onChange={(e) => setNewColor(e.target.value)}
        />
        <Button type="submit" loading={creating}>
          新增标签
        </Button>
      </form>
      {loading && <p>标签加载中...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && (
        <table className="admin-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>Slug</th>
              <th>颜色</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {tags.map((t) => (
              <tr key={t.id}>
                <td>{t.name}</td>
                <td>{t.slug}</td>
                <td>
                  {t.color || "-"}
                  {t.color && (
                    <span
                      style={{
                        display: "inline-block",
                        width: 12,
                        height: 12,
                        borderRadius: "50%",
                        backgroundColor: t.color,
                        marginLeft: 6,
                        border: "1px solid #ddd"
                      }}
                    ></span>
                  )}
                </td>
                <td>
                  <Button
                    type="button"
                    variant="danger"
                    size="small"
                    onClick={() => handleDelete(t.id)}
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


