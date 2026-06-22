import { Navigate, Outlet } from "react-router-dom";
import { useFleet } from "@/lib/fleet-context";

export function ProtectedRoute() {
  const { auth } = useFleet();
  if (!auth) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

export function GuestRoute() {
  const { auth } = useFleet();
  if (auth) {
    return <Navigate to="/app" replace />;
  }
  return <Outlet />;
}
