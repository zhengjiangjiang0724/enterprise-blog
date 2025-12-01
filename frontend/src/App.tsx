import { Route, Routes, Navigate, Link } from "react-router-dom";
import { ArticleList } from "./components/ArticleList";
import { ArticleDetail } from "./components/ArticleDetail";
import { ArticleEditor } from "./components/ArticleEditor";
import { Login } from "./components/Login";
import { Register } from "./components/Register";
import { Profile } from "./components/Profile";
import { AdminUserList } from "./components/AdminUserList";
import { AdminUserDetail } from "./components/AdminUserDetail";
import { AdminArticleList } from "./components/AdminArticleList";
import { AdminArticleDetail } from "./components/AdminArticleDetail";
import { AdminDashboard } from "./components/AdminDashboard";
import { AdminSettings } from "./components/AdminSettings";
import { AdminCategoryList } from "./components/AdminCategoryList";
import { AdminTagList } from "./components/AdminTagList";
import { useAuth } from "./hooks/useAuth";
import { RequireAuth, RequireRole } from "./components/RouteGuards";

export default function App() {
  const { token, user, logout } = useAuth();
  const isAdmin = user?.role === "admin";

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header-left">
          <h1>企业博客系统</h1>
          <nav>
            <Link to="/">文章列表</Link>
            {token && (
              <>
                <Link to="/articles/new">新建文章</Link>
                <Link to="/profile">个人资料</Link>
                {isAdmin && (
                  <>
                    <Link to="/admin/dashboard">仪表盘</Link>
                    <Link to="/admin/articles">文章管理</Link>
                    <Link to="/admin/categories">分类管理</Link>
                    <Link to="/admin/tags">标签管理</Link>
                    <Link to="/admin/users">用户管理</Link>
                    <Link to="/admin/settings">系统配置</Link>
                  </>
                )}
              </>
            )}
          </nav>
        </div>
        <div className="app-header-right">
          {token ? (
            <>
              <span className="user-status">
                {user?.username}（{user?.role || "用户"}）
              </span>
              <button onClick={logout}>退出登录</button>
            </>
          ) : (
            <div className="auth-links">
              <Link to="/login">登录</Link>
              <Link to="/register">注册</Link>
            </div>
          )}
        </div>
      </header>

      <main className="app-main">
        <Routes>
          <Route path="/" element={<ArticleList />} />
          <Route path="/articles/:id" element={<ArticleDetail />} />
          <Route
            path="/articles/new"
            element={
              <RequireAuth>
                <ArticleEditor />
              </RequireAuth>
            }
          />
          <Route
            path="/articles/:id/edit"
            element={
              <RequireAuth>
                <ArticleEditor />
              </RequireAuth>
            }
          />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route
            path="/profile"
            element={
              <RequireAuth>
                <Profile />
              </RequireAuth>
            }
          />
          <Route
            path="/admin/dashboard"
            element={
              <RequireRole roles="admin">
                <AdminDashboard />
              </RequireRole>
            }
          />
          <Route
            path="/admin/users"
            element={
              <RequireRole roles="admin">
                <AdminUserList />
              </RequireRole>
            }
          />
          <Route
            path="/admin/users/:id"
            element={
              <RequireRole roles="admin">
                <AdminUserDetail />
              </RequireRole>
            }
          />
          <Route
            path="/admin/articles"
            element={
              <RequireRole roles="admin">
                <AdminArticleList />
              </RequireRole>
            }
          />
          <Route
            path="/admin/articles/:id"
            element={
              <RequireRole roles="admin">
                <AdminArticleDetail />
              </RequireRole>
            }
          />
          <Route
            path="/admin/categories"
            element={
              <RequireRole roles="admin">
                <AdminCategoryList />
              </RequireRole>
            }
          />
          <Route
            path="/admin/tags"
            element={
              <RequireRole roles="admin">
                <AdminTagList />
              </RequireRole>
            }
          />
          <Route
            path="/admin/settings"
            element={
              <RequireRole roles="admin">
                <AdminSettings />
              </RequireRole>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}


