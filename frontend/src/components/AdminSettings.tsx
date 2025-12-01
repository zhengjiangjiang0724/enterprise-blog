import { useEffect, useState } from "react";
import { apiClient } from "../api/client";
import type { ApiResponse, SystemConfigInfo } from "../api/types";
import { useAuth } from "../hooks/useAuth";

export function AdminSettings() {
  const { token, user } = useAuth();
  const [config, setConfig] = useState<SystemConfigInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isAdmin = user?.role === "admin";

  useEffect(() => {
    if (!token || !isAdmin) return;
    async function fetchConfig() {
      setLoading(true);
      setError(null);
      try {
        const res = await apiClient.get<ApiResponse<SystemConfigInfo>>(
          "/admin/system/config"
        );
        if (res.data.code !== 200 || !res.data.data) {
          throw new Error(res.data.message || "系统配置加载失败");
        }
        setConfig(res.data.data);
      } catch (e: any) {
        setError(e.message || "发生未知错误");
      } finally {
        setLoading(false);
      }
    }
    fetchConfig();
  }, [token, isAdmin]);

  if (!token) {
    return <p className="error">请登录管理员账号。</p>;
  }
  if (!isAdmin) {
    return <p className="error">无权限访问，仅管理员可用。</p>;
  }

  return (
    <div className="admin-section">
      <h2>系统配置（只读）</h2>
      <p className="meta">
        当前版本先以只读形式展示关键配置，如需在线修改可在后续迭代中接入数据库持久化和动态重载。
      </p>
      {loading && <p>配置加载中...</p>}
      {error && <p className="error">{error}</p>}
      {config && !loading && !error && (
        <div className="admin-card config-grid">
          <section>
            <h3>服务器</h3>
            <p>Host：{config.server.host}</p>
            <p>Port：{config.server.port}</p>
            <p>Mode：{config.server.mode}</p>
          </section>
          <section>
            <h3>数据库</h3>
            <p>
              地址：{config.database.host}:{config.database.port}
            </p>
            <p>库名：{config.database.name}</p>
            <p>用户：{config.database.user}</p>
            <p>Max Open Conns：{config.database.max_open_conns}</p>
            <p>Max Idle Conns：{config.database.max_idle_conns}</p>
            <p>
              Conn Lifetime：
              {config.database.conn_max_lifetime_minutes} 分钟
            </p>
          </section>
          <section>
            <h3>Redis</h3>
            <p>
              地址：{config.redis.host}:{config.redis.port}
            </p>
            <p>DB：{config.redis.db}</p>
          </section>
          <section>
            <h3>JWT</h3>
            <p>过期时间：{config.jwt.expire_hours} 小时</p>
          </section>
          <section>
            <h3>日志</h3>
            <p>Level：{config.log.level}</p>
            <p>文件：{config.log.file || "stdout"}</p>
          </section>
          <section>
            <h3>上传配置</h3>
            <p>目录：{config.upload.dir}</p>
            <p>最大尺寸：{(config.upload.max_size / (1024 * 1024)).toFixed(1)} MB</p>
            <p>允许扩展名：{config.upload.exts.join(", ")}</p>
          </section>
        </div>
      )}
    </div>
  );
}


