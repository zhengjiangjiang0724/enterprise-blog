import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, Article, ArticlePayload } from "../api/types";
import { useAuth } from "../hooks/useAuth";

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

  const [form, setForm] = useState(emptyForm);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setError("You must login to create or edit articles.");
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
          throw new Error(res.data.message || "Failed to load article");
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
        setError(e.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchArticle();
  }, [isEdit, id]);

  const handleChange = (field: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
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
    status: form.status,
    category_id: form.category_id || null,
    tag_ids: parseTagIds()
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) {
      setError("You must login to create or edit articles.");
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
        throw new Error(res.data.message || "Operation failed");
      }
      navigate(`/articles/${res.data.data.id}`);
    } catch (e: any) {
      setError(e.response?.data?.message || e.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="article-form">
        <p className="error">You must login to access this page.</p>
      </div>
    );
  }

  return (
    <div className="article-form">
      <h2>{isEdit ? "Edit Article" : "Create Article"}</h2>
      <form onSubmit={handleSubmit}>
        <label>
          Title
          <input
            value={form.title}
            onChange={handleChange("title")}
            required
          />
        </label>
        <label>
          Excerpt
          <input value={form.excerpt} onChange={handleChange("excerpt")} />
        </label>
        <label>
          Cover Image URL
          <input
            value={form.cover_image}
            onChange={handleChange("cover_image")}
          />
        </label>
        <label>
          Status
          <select value={form.status} onChange={handleChange("status")}>
            {STATUS_OPTIONS.map((status) => (
              <option value={status} key={status}>
                {status}
              </option>
            ))}
          </select>
        </label>
        <label>
          Category ID
          <input
            value={form.category_id}
            onChange={handleChange("category_id")}
            placeholder="Optional UUID"
          />
        </label>
        <label>
          Tag IDs (comma separated UUIDs)
          <input
            value={form.tag_ids_text}
            onChange={handleChange("tag_ids_text")}
            placeholder="id1,id2,id3"
          />
        </label>
        <label>
          Content
          <textarea
            value={form.content}
            onChange={handleChange("content")}
            rows={10}
            required
          />
        </label>
        {error && <p className="error">{error}</p>}
        <button type="submit" className="button primary" disabled={loading}>
          {loading ? "Saving..." : isEdit ? "Update Article" : "Create Article"}
        </button>
      </form>
    </div>
  );
}

