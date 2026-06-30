import type { HttpTransport } from "../http.js";
import type { Account, Session, TokenBundle } from "../types.js";

export class AccountService {
  constructor(private readonly http: HttpTransport) {}

  async signUp(input: {
    email: string;
    password: string;
    name: string;
  }): Promise<{ account: Account; tokens: TokenBundle }> {
    const res = await this.http.request<{ account: Account; tokens: TokenBundle }>(
      "POST",
      "/v1/account/sign-up",
      {
        auth: "none",
        body: {
          project_id: this.http.getProjectId(),
          email: input.email,
          password: input.password,
          name: input.name,
        },
      }
    );
    this.http.setAccessToken(res.tokens.access_token);
    return res;
  }

  async signIn(input: {
    email: string;
    password: string;
  }): Promise<{ account: Account; tokens: TokenBundle }> {
    const res = await this.http.request<{ account: Account; tokens: TokenBundle }>(
      "POST",
      "/v1/account/sign-in",
      {
        auth: "none",
        body: {
          project_id: this.http.getProjectId(),
          email: input.email,
          password: input.password,
        },
      }
    );
    this.http.setAccessToken(res.tokens.access_token);
    return res;
  }

  async signOut(): Promise<void> {
    await this.http.request<void>("POST", "/v1/account/sign-out", {
      body: { project_id: this.http.getProjectId() },
    });
    this.http.setAccessToken(undefined);
  }

  async refresh(refreshToken: string): Promise<TokenBundle> {
    const res = await this.http.request<{ tokens: TokenBundle }>("POST", "/v1/account/refresh", {
      auth: "none",
      body: {
        project_id: this.http.getProjectId(),
        refresh_token: refreshToken,
      },
    });
    this.http.setAccessToken(res.tokens.access_token);
    return res.tokens;
  }

  async me(): Promise<Account> {
    return this.http.request<Account>("GET", "/v1/account/me", {
      query: { project_id: this.http.getProjectId() },
    });
  }

  async updateAccount(input: {
    name?: string;
    email?: string;
    password?: string;
    old_password?: string;
  }): Promise<Account> {
    return this.http.request<Account>("PATCH", "/v1/account", { body: input });
  }

  async listSessions(): Promise<Session[]> {
    const res = await this.http.request<{ sessions: Session[] }>("GET", "/v1/account/sessions");
    return res.sessions ?? [];
  }

  async deleteSession(sessionId: string): Promise<void> {
    await this.http.request<void>("DELETE", `/v1/account/sessions/${sessionId}`);
  }

  async deleteSessions(keepCurrent = false): Promise<void> {
    await this.http.request<void>("DELETE", "/v1/account/sessions", {
      body: { keep_current: keepCurrent },
    });
  }

  async getPrefs(): Promise<Record<string, unknown>> {
    const res = await this.http.request<{ prefs: Record<string, unknown> }>(
      "GET",
      "/v1/account/prefs"
    );
    return res.prefs ?? {};
  }

  async updatePrefs(prefs: Record<string, unknown>): Promise<Record<string, unknown>> {
    const res = await this.http.request<{ prefs: Record<string, unknown> }>(
      "PUT",
      "/v1/account/prefs",
      { body: { prefs } }
    );
    return res.prefs ?? {};
  }

  async createOAuth2Session(input: {
    provider: string;
    success: string;
    failure: string;
  }): Promise<{ redirect_url: string }> {
    return this.http.request("GET", `/v1/account/sessions/oauth2/${encodeURIComponent(input.provider)}`, {
      auth: "none",
      query: {
        project_id: this.http.getProjectId(),
        success: input.success,
        failure: input.failure,
      },
    });
  }

  async createOAuth2TokenSession(input: {
    provider: string;
    code: string;
    state: string;
    success?: string;
    failure?: string;
  }): Promise<{ account: Account; tokens: TokenBundle }> {
    const res = await this.http.request<{ account: Account; tokens: TokenBundle }>(
      "POST",
      `/v1/account/sessions/oauth2/${encodeURIComponent(input.provider)}/token`,
      {
        auth: "none",
        body: {
          project_id: this.http.getProjectId(),
          code: input.code,
          state: input.state,
          success: input.success,
          failure: input.failure,
        },
      }
    );
    this.http.setAccessToken(res.tokens.access_token);
    return res;
  }

  async createEmailOTP(input: { email: string }): Promise<{ challenge_id: string; expire_at: number }> {
    return this.http.request("POST", "/v1/account/sessions/email-otp", {
      auth: "none",
      body: {
        project_id: this.http.getProjectId(),
        email: input.email,
      },
    });
  }

  async createEmailOTPSession(input: {
    email: string;
    challenge_id: string;
    otp: string;
  }): Promise<{ account: Account; tokens: TokenBundle }> {
    const res = await this.http.request<{ account: Account; tokens: TokenBundle }>(
      "POST",
      "/v1/account/sessions/email-otp/verify",
      {
        auth: "none",
        body: {
          project_id: this.http.getProjectId(),
          email: input.email,
          challenge_id: input.challenge_id,
          otp: input.otp,
        },
      }
    );
    this.http.setAccessToken(res.tokens.access_token);
    return res;
  }
}
