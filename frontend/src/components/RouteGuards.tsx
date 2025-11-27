import { Navigate, useLocation } from "react-router-dom";
import { ReactNode } from "react";
import { useAuth } from "../hooks/useAuth";

interface RequireAuthProps {
  children: ReactNode;
}

export function RequireAuth({ children }: RequireAuthProps) {
  const { token } = useAuth();
  const location = useLocation();

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <>{children}</>;
}

interface RequireRoleProps {
  children: ReactNode;
  roles: string[] | string;
}

export function RequireRole({ children, roles }: RequireRoleProps) {
  const { token, user } = useAuth();
  const location = useLocation();
  const roleList = Array.isArray(roles) ? roles : [roles];

  if (!token || !user) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (!roleList.includes(user.role)) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}


