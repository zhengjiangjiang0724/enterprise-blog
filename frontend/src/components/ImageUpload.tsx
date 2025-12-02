import { useState, useRef } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, Image } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { BackButton } from "./BackButton";
import { useMessage } from "./MessageProvider";

/**
 * @component ImageUpload
 * @description 图片上传组件，支持拖拽上传和点击上传
 */
export function ImageUpload() {
  const { token } = useAuth();
  const { showSuccess, showError } = useMessage();
  const [uploading, setUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [description, setDescription] = useState("");
  const [tags, setTags] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!token) {
    return <p className="error">请先登录后再上传图片。</p>;
  }

  const handleFileSelect = (file: File) => {
    // 验证文件类型
    const allowedTypes = ["image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"];
    if (!allowedTypes.includes(file.type)) {
      showError("不支持的图片格式，仅支持 JPEG、PNG、GIF、WebP");
      return;
    }

    // 验证文件大小（10MB）
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
      showError("图片大小不能超过 10MB");
      return;
    }

    // 显示预览
    const reader = new FileReader();
    reader.onload = (e) => {
      setPreview(e.target?.result as string);
    };
    reader.readAsDataURL(file);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
  };

  const handleUpload = async () => {
    const fileInput = fileInputRef.current;
    if (!fileInput || !fileInput.files || !fileInput.files[0]) {
      showError("请选择要上传的图片");
      return;
    }

    const file = fileInput.files[0];
    setUploading(true);

    try {
      const formData = new FormData();
      formData.append("file", file);
      if (description.trim()) {
        formData.append("description", description.trim());
      }
      if (tags.trim()) {
        formData.append("tags", tags.trim());
      }

      const res = await apiClient.post<ApiResponse<Image>>(
        "/images/upload",
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data"
          }
        }
      );

      if (res.data.code !== 200 && res.data.code !== 201) {
        throw new Error(res.data.message || "上传失败");
      }

      showSuccess("图片上传成功");
      // 重置表单
      setPreview(null);
      setDescription("");
      setTags("");
      if (fileInput) {
        fileInput.value = "";
      }
    } catch (e: any) {
      const msg = e.response?.data?.message || e.message || "上传失败";
      showError(msg);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="image-upload">
      <div style={{ marginBottom: "16px" }}>
        <BackButton to="/images" label="返回图片管理" />
      </div>
      <h2>上传图片</h2>
      <div className="upload-area">
        <div
          className="drop-zone"
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onClick={() => fileInputRef.current?.click()}
        >
          {preview ? (
            <div className="preview-container">
              <img src={preview} alt="预览" className="preview-image" />
              <button
                className="remove-preview"
                onClick={(e) => {
                  e.stopPropagation();
                  setPreview(null);
                  if (fileInputRef.current) {
                    fileInputRef.current.value = "";
                  }
                }}
              >
                ×
              </button>
            </div>
          ) : (
            <div className="drop-zone-content">
              <p>点击或拖拽图片到这里上传</p>
              <p className="hint">支持 JPEG、PNG、GIF、WebP 格式，最大 10MB</p>
            </div>
          )}
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
            onChange={handleFileChange}
            style={{ display: "none" }}
          />
        </div>

        <div className="upload-form">
          <label>
            图片描述（可选）
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="输入图片描述..."
              rows={3}
            />
          </label>
          <label>
            标签（可选，用逗号分隔）
            <input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="例如：风景,旅游,摄影"
            />
          </label>
          <Button onClick={handleUpload} loading={uploading} disabled={!preview}>
            上传图片
          </Button>
        </div>
      </div>
    </div>
  );
}

