import { Navigate, Outlet } from "react-router-dom";
import { useGraviton } from "@/lib/graviton-context";

export function ProtectedRoute() {
  const { auth } = useGraviton();
  if (!auth) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

export function GuestRoute() {
  const { auth } = useGraviton();
  if (auth) {
    return <Navigate to="/app" replace />;
  }
  return <Outlet />;
}
