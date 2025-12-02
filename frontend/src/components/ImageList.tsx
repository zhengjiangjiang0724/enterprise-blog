import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Image, PaginatedResponse } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { BackButton } from "./BackButton";
import { useMessage } from "./MessageProvider";

/**
 * @component ImageList
 * @description 图片管理列表组件，支持查看、删除、搜索图片
 */
export function ImageList() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useMessage();
  const [images, setImages] = useState<Image[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");
  const [uploaderFilter, setUploaderFilter] = useState<string>(""); // "mine" | "all" | ""
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const isAdmin = user?.role === "admin";

  // 获取图片完整URL
  const getImageUrl = (url: string): string => {
    if (url.startsWith("http")) {
      return url;
    }
    const baseURL = import.meta.env.VITE_API_BASE_URL?.replace("/api/v1", "") || "http://localhost:8080";
    return `${baseURL}${url}`;
  };

  const loadImages = async () => {
    if (!token) return;

    setLoading(true);
    setError(null);
    try {
      const params: Record<string, string> = {
        page: page.toString(),
        page_size: pageSize.toString()
      };

      if (search.trim()) {
        params.search = search.trim();
      }

      if (uploaderFilter === "mine" && user?.id) {
        params.uploader_id = user.id;
      }

      const res = await apiClient.get<PaginatedResponse<Image>>("/images", {
        params
      });

      if (res.data.code !== 200) {
        throw new Error(res.data.message || "图片列表加载失败");
      }

      setImages(res.data.data ?? []);
      setTotal(res.data.meta?.total ?? 0);
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "图片列表加载失败";
      setError(msg);
      showError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadImages();
  }, [token, page, search, uploaderFilter]);

  if (!token) {
    return <p className="error">请先登录后再查看图片。</p>;
  }

  const handleDelete = async (id: string) => {
    if (!confirm("确定要删除这张图片吗？此操作不可恢复。")) {
      return;
    }

    setDeletingId(id);
    try {
      const res = await apiClient.delete<ApiResponse<null>>(`/images/${id}`);
      if (res.data.code !== 200) {
        throw new Error(res.data.message || "删除失败");
      }
      showSuccess("图片已删除");
      loadImages();
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "删除失败";
      showError(msg);
    } finally {
      setDeletingId(null);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
    loadImages();
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="image-list">
      <div style={{ marginBottom: "16px", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <BackButton to="/" label="返回首页" />
        <Link to="/images/upload">
          <Button variant="primary">上传图片</Button>
        </Link>
      </div>
      <h2>图片管理</h2>

      <div className="image-list-filters">
        <form onSubmit={handleSearch} className="search-form">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索图片名称或描述..."
          />
          <Button type="submit">搜索</Button>
        </form>

        <div className="filter-buttons">
          <Button
            variant={uploaderFilter === "" ? "primary" : "secondary"}
            onClick={() => {
              setUploaderFilter("");
              setPage(1);
            }}
          >
            全部图片
          </Button>
          <Button
            variant={uploaderFilter === "mine" ? "primary" : "secondary"}
            onClick={() => {
              setUploaderFilter("mine");
              setPage(1);
            }}
          >
            我的图片
          </Button>
        </div>
      </div>

      {error && <p className="error">{error}</p>}

      {loading ? (
        <p>加载中...</p>
      ) : images.length === 0 ? (
        <p>暂无图片</p>
      ) : (
        <>
          <div className="image-grid">
            {images.map((image) => (
              <div key={image.id} className="image-item">
                <div className="image-item-header">
                  <span className="image-filename">{image.original_name}</span>
                  <span className="image-size">
                    {image.width} × {image.height} ({Math.round(image.size / 1024)}KB)
                  </span>
                </div>
                <div className="image-item-preview">
                  <img
                    src={getImageUrl(image.url)}
                    alt={image.description || image.original_name}
                    onError={(e) => {
                      // 如果图片加载失败，显示占位符
                      (e.target as HTMLImageElement).src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='200'%3E%3Crect fill='%23ddd' width='200' height='200'/%3E%3Ctext fill='%23999' font-family='sans-serif' font-size='14' x='50%25' y='50%25' text-anchor='middle' dominant-baseline='middle'%3E图片加载失败%3C/text%3E%3C/svg%3E";
                    }}
                  />
                </div>
                {image.description && (
                  <div className="image-item-description">{image.description}</div>
                )}
                {image.tags && image.tags.length > 0 && (
                  <div className="image-item-tags">
                    {image.tags.map((tag, idx) => (
                      <span key={idx} className="tag">{tag}</span>
                    ))}
                  </div>
                )}
                <div className="image-item-footer">
                  <span className="image-date">
                    {new Date(image.created_at).toLocaleDateString("zh-CN")}
                  </span>
                </div>
                <div className="image-item-actions">
                  <a
                    href={getImageUrl(image.url)}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Button variant="secondary">查看</Button>
                  </a>
                  {(isAdmin || image.uploader_id === user?.id) && (
                    <Button
                      variant="danger"
                      onClick={() => handleDelete(image.id)}
                      loading={deletingId === image.id}
                    >
                      删除
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="pagination">
              <Button
                variant="secondary"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                上一页
              </Button>
              <span>
                第 {page} / {totalPages} 页（共 {total} 张图片）
              </span>
              <Button
                variant="secondary"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
              >
                下一页
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

