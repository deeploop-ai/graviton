import { Fleet, FleetError } from "@fleet/sdk";
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

interface FleetContextValue {
  settings: AppSettings;
  auth: AuthState | null;
  client: Fleet;
  updateSettings: (next: AppSettings) => void;
  setAuth: (next: AuthState | null) => void;
  serverFleet: () => Fleet;
  lastError: string | null;
  setLastError: (msg: string | null) => void;
  run: <T>(fn: () => Promise<T>) => Promise<T>;
}

const FleetContext = createContext<FleetContextValue | null>(null);

function buildClient(settings: AppSettings, auth: AuthState | null): Fleet {
  const fleet = Fleet.create({
    endpoint: settings.endpoint,
    projectId: settings.projectId,
    accessToken: auth?.accessToken,
  });
  return fleet;
}

export function FleetProvider({ children }: { children: ReactNode }) {
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

  const serverFleet = useCallback(() => {
    if (!settings.apiKey) {
      throw new FleetError("请先在设置页填写 Server API Key", 0);
    }
    return Fleet.withApiKey(settings.endpoint, settings.projectId, settings.apiKey);
  }, [settings]);

  const run = useCallback(async <T,>(fn: () => Promise<T>): Promise<T> => {
    setLastError(null);
    try {
      return await fn();
    } catch (err) {
      const message =
        err instanceof FleetError
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
      serverFleet,
      lastError,
      setLastError,
      run,
    }),
    [settings, auth, client, updateSettings, setAuth, serverFleet, lastError, run]
  );

  return <FleetContext.Provider value={value}>{children}</FleetContext.Provider>;
}

export function useFleet() {
  const ctx = useContext(FleetContext);
  if (!ctx) throw new Error("useFleet must be used within FleetProvider");
  return ctx;
}
