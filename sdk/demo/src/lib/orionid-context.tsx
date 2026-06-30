import { Orionid, OrionidError } from "@orionid/sdk";
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

interface OrionidContextValue {
  settings: AppSettings;
  auth: AuthState | null;
  client: Orionid;
  updateSettings: (next: AppSettings) => void;
  setAuth: (next: AuthState | null) => void;
  serverClient: () => Orionid;
  lastError: string | null;
  setLastError: (msg: string | null) => void;
  run: <T>(fn: () => Promise<T>) => Promise<T>;
}

const OrionidContext = createContext<OrionidContextValue | null>(null);

function buildClient(settings: AppSettings, auth: AuthState | null): Orionid {
  return Orionid.create({
    endpoint: settings.endpoint,
    projectId: settings.projectId,
    accessToken: auth?.accessToken,
  });
}

export function OrionidProvider({ children }: { children: ReactNode }) {
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
      throw new OrionidError("请先在设置页填写 Server API Key", 0);
    }
    return Orionid.withApiKey(settings.endpoint, settings.projectId, settings.apiKey);
  }, [settings]);

  const run = useCallback(async <T,>(fn: () => Promise<T>): Promise<T> => {
    setLastError(null);
    try {
      return await fn();
    } catch (err) {
      const message =
        err instanceof OrionidError
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

  return <OrionidContext.Provider value={value}>{children}</OrionidContext.Provider>;
}

export function useOrionid() {
  const ctx = useContext(OrionidContext);
  if (!ctx) throw new Error("useOrionid must be used within OrionidProvider");
  return ctx;
}
