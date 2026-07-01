import { Graviton, GravitonError } from "@graviton/sdk";
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  loadAuth,
  loadSettings,
  saveAuth,
  saveSettings,
  type AppSettings,
  type AuthState,
} from "./storage";

interface GravitonContextValue {
  settings: AppSettings;
  auth: AuthState | null;
  client: Graviton;
  updateSettings: (next: AppSettings) => void;
  setAuth: (next: AuthState | null) => void;
  serverClient: () => Graviton;
  lastError: string | null;
  setLastError: (msg: string | null) => void;
  run: <T>(fn: () => Promise<T>) => Promise<T>;
}

const GravitonContext = createContext<GravitonContextValue | null>(null);

function buildClient(settings: AppSettings, auth: AuthState | null): Graviton {
  return Graviton.create({
    endpoint: settings.endpoint,
    projectId: settings.projectId,
    accessToken: auth?.accessToken,
  });
}

export function GravitonProvider({ children }: { children: ReactNode }) {
  const [settings, setSettingsState] = useState<AppSettings>(() => loadSettings());
  const [auth, setAuthState] = useState<AuthState | null>(() => loadAuth());
  const [lastError, setLastError] = useState<string | null>(null);

  const client = useMemo(() => buildClient(settings, auth), [settings, auth]);

  const updateSettings = useCallback((next: AppSettings) => {
    saveSettings(next);
    setSettingsState(next);
  }, []);

  const setAuth = useCallback((next: AuthState | null) => {
    saveAuth(next);
    setAuthState(next);
    if (next) {
      client.setAccessToken(next.accessToken);
    } else {
      client.setAccessToken(undefined);
    }
  }, [client]);

  const serverClient = useCallback(() => {
    if (!settings.apiKey) {
      throw new GravitonError("请先在设置页填写 Server API Key", 0);
    }
    return Graviton.withApiKey(settings.endpoint, settings.projectId, settings.apiKey);
  }, [settings]);

  const run = useCallback(async <T,>(fn: () => Promise<T>): Promise<T> => {
    setLastError(null);
    try {
      return await fn();
    } catch (err) {
      const message =
        err instanceof GravitonError
          ? `[${err.status}] ${err.message}`
          : err instanceof Error
            ? err.message
            : String(err);
      setLastError(message);
      throw err;
    }
  }, []);

  const value = useMemo(
    () => ({
      settings,
      auth,
      client,
      updateSettings,
      setAuth,
      serverClient,
      lastError,
      setLastError,
      run,
    }),
    [settings, auth, client, updateSettings, setAuth, serverClient, lastError, run]
  );

  return <GravitonContext.Provider value={value}>{children}</GravitonContext.Provider>;
}

export function useGraviton() {
  const ctx = useContext(GravitonContext);
  if (!ctx) throw new Error("useGraviton must be used within GravitonProvider");
  return ctx;
}
