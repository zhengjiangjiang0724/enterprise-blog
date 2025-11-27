import { Route, Routes, Navigate, Link } from "react-router-dom";
import { ArticleList } from "./components/ArticleList";
import { ArticleDetail } from "./components/ArticleDetail";
import { ArticleEditor } from "./components/ArticleEditor";
import { Login } from "./components/Login";
import { Register } from "./components/Register";
import { Profile } from "./components/Profile";
import { AdminUserList } from "./components/AdminUserList";
import { AdminUserDetail } from "./components/AdminUserDetail";
import { useAuth } from "./hooks/useAuth";

export default function App() {
  const { token, user, logout } = useAuth();
  const isAdmin = user?.role === "admin";

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header-left">
          <h1>Enterprise Blog</h1>
          <nav>
            <Link to="/">Articles</Link>
            {token && (
              <>
                <Link to="/articles/new">New Article</Link>
                <Link to="/profile">Profile</Link>
                {isAdmin && <Link to="/admin/users">User Management</Link>}
              </>
            )}
          </nav>
        </div>
        <div className="app-header-right">
          {token ? (
            <>
              <span className="user-status">
                {user?.username} ({user?.role || "user"})
              </span>
              <button onClick={logout}>Logout</button>
            </>
          ) : (
            <div className="auth-links">
              <Link to="/login">Login</Link>
              <Link to="/register">Register</Link>
            </div>
          )}
        </div>
      </header>

      <main className="app-main">
        <Routes>
          <Route path="/" element={<ArticleList />} />
          <Route path="/articles/:id" element={<ArticleDetail />} />
          <Route path="/articles/new" element={<ArticleEditor />} />
          <Route path="/articles/:id/edit" element={<ArticleEditor />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/admin/users" element={<AdminUserList />} />
          <Route path="/admin/users/:id" element={<AdminUserDetail />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}


