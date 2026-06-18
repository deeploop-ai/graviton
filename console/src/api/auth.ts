import { api, setAuthToken } from "./client";

export interface LoginInput {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  expires_at: string;
}

export async function login(input: LoginInput): Promise<string> {
  const res = await api.post<LoginResponse>("/console/auth/sign-in", input);
  setAuthToken(res.data.access_token);
  return res.data.access_token;
}

export function logout() {
  setAuthToken(null);
}
