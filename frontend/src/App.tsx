import { Route, Routes, Navigate, Link } from "react-router-dom";
import { ArticleList } from "./components/ArticleList";
import { ArticleDetail } from "./components/ArticleDetail";
import { Login } from "./components/Login";
import { useAuth } from "./hooks/useAuth";

export default function App() {
  const { token, logout } = useAuth();

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header-left">
          <h1>Enterprise Blog</h1>
          <nav>
            <Link to="/">Articles</Link>
          </nav>
        </div>
        <div className="app-header-right">
          {token ? (
            <button onClick={logout}>Logout</button>
          ) : (
            <Link to="/login">Login</Link>
          )}
        </div>
      </header>

      <main className="app-main">
        <Routes>
          <Route path="/" element={<ArticleList />} />
          <Route path="/articles/:id" element={<ArticleDetail />} />
          <Route path="/login" element={<Login />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}


