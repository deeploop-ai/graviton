import type { HttpTransport } from "../http.js";

export interface OAuthProvider {
  provider: string;
  enabled: boolean;
  client_id: string;
  has_client_secret: boolean;
  scopes: string[];
  created_at?: string;
  updated_at?: string;
}

export class OAuthProvidersService {
  constructor(private readonly http: HttpTransport) {}

  async list(): Promise<OAuthProvider[]> {
    const res = await this.http.request<{ oauth_providers: OAuthProvider[] }>(
      "GET",
      "/v1/server/oauth-providers",
      { auth: "apiKey" }
    );
    return res.oauth_providers ?? [];
  }

  async upsert(input: {
    provider: string;
    enabled: boolean;
    client_id: string;
    client_secret?: string;
    scopes?: string[];
  }): Promise<OAuthProvider> {
    return this.http.request<OAuthProvider>(
      "PUT",
      `/v1/server/oauth-providers/${encodeURIComponent(input.provider)}`,
      {
        auth: "apiKey",
        body: {
          provider: input.provider,
          enabled: input.enabled,
          client_id: input.client_id,
          client_secret: input.client_secret ?? "",
          scopes: input.scopes ?? [],
        },
      }
    );
  }

  async delete(provider: string): Promise<void> {
    await this.http.request<void>(
      "DELETE",
      `/v1/server/oauth-providers/${encodeURIComponent(provider)}`,
      { auth: "apiKey" }
    );
  }
}
