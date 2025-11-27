import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article, ArticlePayload } from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";
import ReactMarkdown from "react-markdown";

const STATUS_OPTIONS = ["draft", "published", "archived"];

const emptyForm = {
  title: "",
  content: "",
  excerpt: "",
  cover_image: "",
  status: "draft",
  category_id: "",
  tag_ids_text: ""
};

export function ArticleEditor() {
  const { id } = useParams<{ id: string }>();
  const isEdit = useMemo(() => Boolean(id), [id]);
  const navigate = useNavigate();
  const { token } = useAuth();
  const { showSuccess, showError } = useMessage();

  const [form, setForm] = useState(emptyForm);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"edit" | "preview">("edit");

  useEffect(() => {
    if (!token) {
      setError("请先登录后再创建或编辑文章。");
    }
  }, [token]);

  useEffect(() => {
    if (!isEdit || !id) return;
    async function fetchArticle() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<Article>>(`/articles/${id}`);
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "文章加载失败");
        }
        const article = res.data.data;
        setForm({
          title: article.title || "",
          content: article.content || "",
          excerpt: article.excerpt || "",
          cover_image: article.cover_image || "",
          status: article.status || "draft",
          category_id: article.category_id || "",
          tag_ids_text: article.tags?.map((t) => t.id).join(",") || ""
        });
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchArticle();
  }, [isEdit, id]);

  const handleChange =
    (field: keyof typeof form) =>
    (
      e: React.ChangeEvent<
        HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
      >
    ) => {
      setForm((prev) => ({
        ...prev,
        [field]: e.target.value
      }));
    };

  const parseTagIds = () => {
    if (!form.tag_ids_text.trim()) return undefined;
    return form.tag_ids_text
      .split(",")
      .map((tag) => tag.trim())
      .filter(Boolean);
  };

  const buildPayload = (): ArticlePayload => ({
    title: form.title,
    content: form.content,
    excerpt: form.excerpt,
    cover_image: form.cover_image,
    status: form.status
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) {
      setError("请先登录后再操作。");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const payload = buildPayload();
      let res;
      if (isEdit && id) {
        res = await apiClient.put<ApiResponse<Article>>(
          `/articles/${id}`,
          payload
        );
      } else {
        res = await apiClient.post<ApiResponse<Article>>("/articles", payload);
      }
      if (res.data.code !== 200 || !res.data.data) {
        throw new Error(res.data.message || "操作失败");
      }
      showSuccess(isEdit ? "文章已更新" : "文章已创建");
      navigate(`/articles/${res.data.data.id}`);
    } catch (e: any) {
      const msg =
        e.response?.data?.message || e.message || "发生未知错误";
      setError(msg);
      showError(msg);
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="article-form">
        <p className="error">请先登录后再访问该页面。</p>
      </div>
    );
  }

  return (
    <div className="article-form">
      <h2>{isEdit ? "编辑文章" : "创建文章"}</h2>
      <form onSubmit={handleSubmit}>
        <label>
          标题
          <input
            value={form.title}
            onChange={handleChange("title")}
            required
          />
        </label>
        <label>
          摘要
          <input value={form.excerpt} onChange={handleChange("excerpt")} />
        </label>
        <label>
          封面图片地址
          <input
            value={form.cover_image}
            onChange={handleChange("cover_image")}
          />
        </label>
        <label>
          发布状态
          <select value={form.status} onChange={handleChange("status")}>
            {STATUS_OPTIONS.map((status) => (
              <option value={status} key={status}>
                {status}
              </option>
            ))}
          </select>
        </label>
        {/* 分类和标签功能已简化，暂不在表单中配置 */}
        <div className="editor-section">
          <div className="editor-header">
            <span>正文内容</span>
            <div className="editor-tabs">
              <button
                type="button"
                className={`editor-tab ${
                  activeTab === "edit" ? "active" : ""
                }`}
                onClick={() => setActiveTab("edit")}
              >
                编辑
              </button>
              <button
                type="button"
                className={`editor-tab ${
                  activeTab === "preview" ? "active" : ""
                }`}
                onClick={() => setActiveTab("preview")}
              >
                预览
              </button>
            </div>
          </div>
          <div className="editor-body">
            <textarea
              value={form.content}
              onChange={handleChange("content")}
              rows={10}
              required
            />
            {activeTab === "preview" && (
              <div className="markdown-preview">
                <ReactMarkdown>
                  {form.content || "（暂无内容）"}
                </ReactMarkdown>
              </div>
            )}
          </div>
        </div>
        {error && <p className="error">{error}</p>}
        <div className="editor-actions">
          <Button
            type="submit"
            variant="secondary"
            loading={loading}
            onClick={() =>
              setForm((prev) => ({ ...prev, status: "draft" }))
            }
          >
            保存为草稿
          </Button>
          <Button
            type="submit"
            variant="primary"
            loading={loading}
            style={{ marginLeft: "8px" }}
            onClick={() =>
              setForm((prev) => ({ ...prev, status: "published" }))
            }
          >
            {isEdit ? "更新并发布" : "发布文章"}
          </Button>
        </div>
      </form>
    </div>
  );
}

