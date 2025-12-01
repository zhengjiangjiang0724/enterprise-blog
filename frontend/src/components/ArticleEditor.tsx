import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type {
  ApiResponse,
  Article,
  ArticlePayload,
  Category,
  Tag
} from "../api/types";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";
import { useMessage } from "./MessageProvider";
import ReactMarkdown from "react-markdown";

const STATUS_OPTIONS = ["draft", "review", "published", "archived"];

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
  const { token, user } = useAuth();
  const { showSuccess, showError } = useMessage();

  const [form, setForm] = useState(emptyForm);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"edit" | "preview">("edit");
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!token) {
      setError("请先登录后再创建或编辑文章。");
    }
  }, [token]);

  // 加载分类和标签
  useEffect(() => {
    async function fetchMeta() {
      try {
        const [catRes, tagRes] = await Promise.all([
          apiClient.get<ApiResponse<Category[]>>("/categories"),
          apiClient.get<ApiResponse<Tag[]>>("/tags")
        ]);
        if (catRes.data.code === 200 && Array.isArray(catRes.data.data)) {
          setCategories(catRes.data.data);
        }
        if (tagRes.data.code === 200 && Array.isArray(tagRes.data.data)) {
          setTags(tagRes.data.data);
        }
      } catch {
        // 静默失败，避免影响文章编辑
      }
    }
    fetchMeta();
  }, []);

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
          category_id: article.category?.id || "",
          tag_ids_text: ""
        });
        setSelectedTagIds(article.tags?.map((t) => t.id) || []);
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

  const buildPayload = (): ArticlePayload => ({
    title: form.title,
    content: form.content,
    excerpt: form.excerpt,
    cover_image: form.cover_image,
    status: form.status,
    category_id: form.category_id || undefined,
    tag_ids: selectedTagIds.length > 0 ? selectedTagIds : undefined
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
        {isEdit && (
          <label>
            发布状态
            <select value={form.status} onChange={handleChange("status")}>
              {STATUS_OPTIONS.map((status) => (
                <option value={status} key={status}>
                  {status === "draft"
                    ? "草稿"
                    : status === "review"
                    ? "待审核"
                    : status === "published"
                    ? "已发布"
                    : "已归档"}
                </option>
              ))}
            </select>
          </label>
        )}
        {/* 分类选择 */}
        <label>
          分类
          <select
            value={form.category_id}
            onChange={handleChange("category_id")}
          >
            <option value="">未选择</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
        {/* 标签多选 */}
        <div className="tag-selector">
          <span>标签（可多选）</span>
          <div className="tag-list">
            {tags.map((t) => {
              const checked = selectedTagIds.includes(t.id);
              return (
                <label key={t.id} className="tag-item">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={(e) => {
                      const { checked } = e.target;
                      setSelectedTagIds((prev) =>
                        checked
                          ? [...prev, t.id]
                          : prev.filter((id) => id !== t.id)
                      );
                    }}
                  />
                  <span>{t.name}</span>
                </label>
              );
            })}
            {tags.length === 0 && <span className="meta">暂无标签，可在后台创建。</span>}
          </div>
        </div>
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
            variant="secondary"
            loading={loading}
            style={{ marginLeft: "8px" }}
            onClick={() =>
              setForm((prev) => ({ ...prev, status: "review" }))
            }
          >
            提交审核
          </Button>
          <Button
            type="submit"
            variant="primary"
            loading={loading}
            style={{ marginLeft: "8px" }}
            onClick={() =>
              setForm((prev) => ({
                ...prev,
                status: isAdmin ? "published" : "review"
              }))
            }
          >
            {isAdmin
              ? isEdit
                ? "更新并发布"
                : "直接发布"
              : isEdit
              ? "提交审核（更新）"
              : "提交审核并创建"}
          </Button>
        </div>
      </form>
    </div>
  );
}

