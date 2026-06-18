import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { getAuthToken, clearAuthToken, setProjectID } from "@/api/client";
import { login as apiLogin } from "@/api/auth";

interface AuthContextValue {
  token: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  selectProject: (projectID: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(getAuthToken());

  useEffect(() => {
    setToken(getAuthToken());
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const t = await apiLogin({ email, password });
    setToken(t);
  }, []);

  const logout = useCallback(() => {
    clearAuthToken();
    setProjectID(null);
    setToken(null);
  }, []);

  const selectProject = useCallback((projectID: string) => {
    setProjectID(projectID);
  }, []);

  const value = useMemo(
    () => ({
      token,
      isAuthenticated: !!token,
      login,
      logout,
      selectProject,
    }),
    [token, login, logout, selectProject]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
