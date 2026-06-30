import { api } from "./client";

export interface OAuthProvider {
  provider: string;
  enabled: boolean;
  client_id: string;
  has_client_secret: boolean;
  scopes: string[];
  created_at?: string;
  updated_at?: string;
}

export interface ListOAuthProvidersResponse {
  oauth_providers: OAuthProvider[];
}

export const OAUTH_PROVIDER_OPTIONS = [
  { id: "google", label: "Google", defaultScopes: ["openid", "email", "profile"] },
  { id: "github", label: "GitHub", defaultScopes: ["read:user", "user:email"] },
  { id: "wechat_web", label: "微信 · 网站扫码", defaultScopes: [] },
  { id: "wechat_mp", label: "微信 · 公众号 H5", defaultScopes: [] },
  { id: "wechat_miniprogram", label: "微信 · 小程序", defaultScopes: [] },
] as const;

export async function listOAuthProviders(): Promise<OAuthProvider[]> {
  const res = await api.get<ListOAuthProvidersResponse>("/server/oauth-providers");
  return res.data.oauth_providers ?? [];
}

export async function upsertOAuthProvider(input: {
  provider: string;
  enabled: boolean;
  client_id: string;
  client_secret?: string;
  scopes?: string[];
}): Promise<OAuthProvider> {
  const res = await api.put<OAuthProvider>(
    `/server/oauth-providers/${encodeURIComponent(input.provider)}`,
    {
      provider: input.provider,
      enabled: input.enabled,
      client_id: input.client_id,
      client_secret: input.client_secret ?? "",
      scopes: input.scopes ?? [],
    }
  );
  return res.data;
}

export async function deleteOAuthProvider(provider: string): Promise<void> {
  await api.delete(`/server/oauth-providers/${encodeURIComponent(provider)}`);
}

export function oauthCallbackURL(provider: string, publicBase = window.location.origin): string {
  return `${publicBase.replace(/\/$/, "")}/v1/account/oauth2/${provider}/callback`;
}
