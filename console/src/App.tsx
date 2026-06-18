import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { AuthProvider, useAuth } from "@/hooks/useAuth";
import { Layout } from "@/components/Layout";
import { Login } from "@/routes/Login";
import { Dashboard } from "@/routes/Dashboard";
import { Projects } from "@/routes/Projects";
import { ApiKeys } from "@/routes/ApiKeys";
import { Users } from "@/routes/Users";
import { Storage } from "@/routes/Storage";
import { Databases } from "@/routes/Databases";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) {
    return <Navigate to="/console/login" replace />;
  }
  return <>{children}</>;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/console/login" element={<Login />} />
      <Route
        path="/console"
        element={
          <RequireAuth>
            <Layout />
          </RequireAuth>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="projects" element={<Projects />} />
        <Route path="api-keys" element={<ApiKeys />} />
        <Route path="users" element={<Users />} />
        <Route path="storage" element={<Storage />} />
        <Route path="databases" element={<Databases />} />
      </Route>
      <Route path="*" element={<Navigate to="/console" replace />} />
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
        <Toaster position="top-right" />
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
