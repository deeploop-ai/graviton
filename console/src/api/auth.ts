import { api, setAuthToken } from "./client";

export interface LoginInput {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  expires_at: string;
  refresh_token?: string;
}

export async function login(input: LoginInput): Promise<string> {
  const res = await api.post<LoginResponse>("/console/auth/sign-in", input);
  setAuthToken(res.data.access_token);
  if (res.data.refresh_token) {
    localStorage.setItem("graviton_console_refresh_token", res.data.refresh_token);
  }
  return res.data.access_token;
}

export async function refreshAuthToken(): Promise<string> {
  const refreshToken = localStorage.getItem("graviton_console_refresh_token");
  if (!refreshToken) {
    throw new Error("no refresh token");
  }
  const res = await api.post<LoginResponse>("/console/auth/refresh", {
    refresh_token: refreshToken,
  });
  setAuthToken(res.data.access_token);
  if (res.data.refresh_token) {
    localStorage.setItem("graviton_console_refresh_token", res.data.refresh_token);
  }
  return res.data.access_token;
}

export function logout() {
  setAuthToken(null);
  localStorage.removeItem("graviton_console_refresh_token");
}
