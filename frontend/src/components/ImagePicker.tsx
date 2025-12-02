import { useCallback, useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, Image, PaginatedResponse } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";

/**
 * @interface ImagePickerProps
 * @description 图片选择器组件的属性
 */
interface ImagePickerProps {
  /** 当前选中的图片URL */
  value?: string;
  /** 选择图片时的回调函数 */
  onChange: (url: string) => void;
  /** 是否显示选择器 */
  open: boolean;
  /** 关闭选择器的回调函数 */
  onClose: () => void;
}

/**
 * @component ImagePicker
 * @description 图片选择器组件，用于从已上传的图片库中选择图片
 */
export function ImagePicker({ value, onChange, open, onClose }: ImagePickerProps) {
  const { token } = useAuth();
  const { showError } = useMessage();
  const [images, setImages] = useState<Image[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");

  const loadImages = useCallback(async () => {
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
  }, [token, page, pageSize, search, showError]);

  // 获取图片完整URL
  const getImageUrl = (url: string): string => {
    if (url.startsWith("http")) {
      return url;
    }
    const baseURL = import.meta.env.VITE_API_BASE_URL?.replace("/api/v1", "") || "http://localhost:8080";
    return `${baseURL}${url}`;
  };

  useEffect(() => {
    if (open && token) {
      loadImages();
    }
  }, [open, token, loadImages]);

  const handleSelect = (image: Image) => {
    onChange(getImageUrl(image.url));
    onClose();
  };

  if (!open) return null;

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="image-picker-overlay" onClick={onClose}>
      <div className="image-picker-modal" onClick={(e) => e.stopPropagation()}>
        <div className="image-picker-header">
          <h3>选择图片</h3>
          <button className="image-picker-close" onClick={onClose}>×</button>
        </div>

        <div className="image-picker-search">
          <input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            placeholder="搜索图片..."
          />
        </div>

        {error && <p className="error">{error}</p>}

        {loading ? (
          <p>加载中...</p>
        ) : images.length === 0 ? (
          <p>暂无图片</p>
        ) : (
          <>
            <div className="image-picker-grid">
              {images.map((image) => (
                <div
                  key={image.id}
                  className={`image-picker-item ${value === getImageUrl(image.url) ? "selected" : ""}`}
                  onClick={() => handleSelect(image)}
                >
                  <img
                    src={getImageUrl(image.url)}
                    alt={image.description || image.original_name}
                    onError={(e) => {
                      (e.target as HTMLImageElement).src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='200'%3E%3Crect fill='%23ddd' width='200' height='200'/%3E%3Ctext fill='%23999' font-family='sans-serif' font-size='14' x='50%25' y='50%25' text-anchor='middle' dominant-baseline='middle'%3E图片加载失败%3C/text%3E%3C/svg%3E";
                    }}
                  />
                  {value === getImageUrl(image.url) && (
                    <div className="image-picker-selected-badge">✓</div>
                  )}
                </div>
              ))}
            </div>

            {totalPages > 1 && (
              <div className="image-picker-pagination">
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
    </div>
  );
}

