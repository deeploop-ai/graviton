import { Navigate, Outlet } from "react-router-dom";
import { useOrionid } from "@/lib/orionid-context";

export function ProtectedRoute() {
  const { auth } = useOrionid();
  if (!auth) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

export function GuestRoute() {
  const { auth } = useOrionid();
  if (auth) {
    return <Navigate to="/app" replace />;
  }
  return <Outlet />;
}
