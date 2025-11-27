import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { apiClient } from "../api/client";
import type { ApiResponse, UserProfile } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function AdminUserDetail() {
  const { id } = useParams<{ id: string }>();
  const { token, user } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!token || !isAdmin || !id) return;
    async function fetchUser() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<UserProfile>>(
          `/admin/users/${id}`
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "Failed to load user");
        }
        setProfile(res.data.data);
      } catch (e: any) {
        setError(e.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }
    fetchUser();
  }, [token, isAdmin, id]);

  if (!token) {
    return <p className="error">Please login as admin.</p>;
  }
  if (!isAdmin) {
    return <p className="error">Access denied. Admin only.</p>;
  }

  return (
    <div className="admin-section">
      <h2>User Detail</h2>
      {loading && <p>Loading user...</p>}
      {error && <p className="error">{error}</p>}
      {profile && (
        <div className="admin-card">
          <p>
            <strong>Username:</strong> {profile.username}
          </p>
          <p>
            <strong>Email:</strong> {profile.email}
          </p>
          <p>
            <strong>Role:</strong> {profile.role}
          </p>
          <p>
            <strong>Status:</strong> {profile.status}
          </p>
          <p>
            <strong>Bio:</strong> {profile.bio || "-"}
          </p>
          <p>
            <strong>Created:</strong>{" "}
            {new Date(profile.created_at).toLocaleString()}
          </p>
          <p>
            <strong>Updated:</strong>{" "}
            {new Date(profile.updated_at).toLocaleString()}
          </p>
        </div>
      )}
    </div>
  );
}


