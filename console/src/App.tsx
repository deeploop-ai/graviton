import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { AuthProvider, useAuth } from "@/hooks/useAuth";
import { Layout } from "@/components/Layout";
import { Login } from "@/routes/Login";
import { Dashboard } from "@/routes/Dashboard";
import {
  ProjectsListPage,
  ProjectNewPage,
  ProjectDetailPage,
} from "@/routes/projects/pages";
import {
  ApiKeysListPage,
  ApiKeyNewPage,
  ApiKeyDetailPage,
} from "@/routes/api-keys/pages";
import {
  UsersListPage,
  UserDetailPage,
  UserEditPage,
} from "@/routes/users/pages";
import {
  StorageListPage,
  BucketNewPage,
  BucketDetailPage,
  FileDetailPage,
} from "@/routes/storage/pages";
import {
  TeamsListPage,
  TeamNewPage,
  TeamDetailPage,
} from "@/routes/teams/pages";
import {
  DatabasesListPage,
  DatabaseNewPage,
  DatabaseDetailPage,
  CollectionNewPage,
  CollectionDetailPage,
  DocumentNewPage,
  DocumentDetailPage,
} from "@/routes/databases/pages";

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

        <Route path="projects" element={<ProjectsListPage />} />
        <Route path="projects/new" element={<ProjectNewPage />} />
        <Route path="projects/:id" element={<ProjectDetailPage />} />

        <Route path="api-keys" element={<ApiKeysListPage />} />
        <Route path="api-keys/new" element={<ApiKeyNewPage />} />
        <Route path="api-keys/:id" element={<ApiKeyDetailPage />} />

        <Route path="users" element={<UsersListPage />} />
        <Route path="users/:id" element={<UserDetailPage />} />
        <Route path="users/:id/edit" element={<UserEditPage />} />

        <Route path="teams" element={<TeamsListPage />} />
        <Route path="teams/new" element={<TeamNewPage />} />
        <Route path="teams/:id" element={<TeamDetailPage />} />

        <Route path="storage" element={<StorageListPage />} />
        <Route path="storage/new" element={<BucketNewPage />} />
        <Route path="storage/:bucketId" element={<BucketDetailPage />} />
        <Route path="storage/:bucketId/files/:fileId" element={<FileDetailPage />} />

        <Route path="databases" element={<DatabasesListPage />} />
        <Route path="databases/new" element={<DatabaseNewPage />} />
        <Route path="databases/:dbId" element={<DatabaseDetailPage />} />
        <Route path="databases/:dbId/collections/new" element={<CollectionNewPage />} />
        <Route path="databases/:dbId/collections/:collId" element={<CollectionDetailPage />} />
        <Route path="databases/:dbId/collections/:collId/documents/new" element={<DocumentNewPage />} />
        <Route path="databases/:dbId/collections/:collId/documents/:docId" element={<DocumentDetailPage />} />
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
