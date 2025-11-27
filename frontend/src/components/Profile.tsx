import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, UserProfile } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function Profile() {
  const { token } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [form, setForm] = useState({
    username: "",
    email: "",
    avatar: "",
    bio: ""
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (!token) return;
    async function fetchProfile() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<UserProfile>>(
          "/users/profile"
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "Failed to load profile");
        }
        setProfile(res.data.data);
        setForm({
          username: res.data.data.username || "",
          email: res.data.data.email || "",
          avatar: res.data.data.avatar || "",
          bio: res.data.data.bio || ""
        });
      } catch (e: any) {
        setError(e.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchProfile();
  }, [token]);

  const handleChange =
    (field: keyof typeof form) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm((prev) => ({
        ...prev,
        [field]: e.target.value
      }));
    };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      const res = await apiClient.put<ApiResponse<UserProfile>>(
        "/users/profile",
        {
          username: form.username,
          email: form.email,
          avatar: form.avatar,
          bio: form.bio
        }
      );
      if (res.data.code !== 200 || !res.data.data) {
        throw new Error(res.data.message || "Failed to update profile");
      }
      setProfile(res.data.data);
      setSuccess("Profile updated successfully");
    } catch (e: any) {
      setError(e.response?.data?.message || e.message || "Unknown error");
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="profile-form">
        <p className="error">Please login to view your profile.</p>
      </div>
    );
  }

  return (
    <div className="profile-form">
      <h2>Profile</h2>
      {loading && !profile && <p>Loading profile...</p>}
      {error && <p className="error">{error}</p>}
      {success && <p className="success">{success}</p>}
      {profile && (
        <form onSubmit={handleSubmit}>
          <label>
            Username
            <input
              value={form.username}
              onChange={handleChange("username")}
              required
            />
          </label>
          <label>
            Email
            <input
              type="email"
              value={form.email}
              onChange={handleChange("email")}
              required
            />
          </label>
          <label>
            Avatar URL
            <input value={form.avatar} onChange={handleChange("avatar")} />
          </label>
          <label>
            Bio
            <textarea
              rows={4}
              value={form.bio}
              onChange={handleChange("bio")}
            />
          </label>
          <button type="submit" className="button primary" disabled={loading}>
            {loading ? "Saving..." : "Update Profile"}
          </button>
        </form>
      )}
    </div>
  );
}

