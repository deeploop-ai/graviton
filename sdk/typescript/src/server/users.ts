import { listQuery, type HttpTransport } from "../http.js";
import type { ListParams, User } from "../types.js";

export class UsersService {
  constructor(private readonly http: HttpTransport) {}

  async list(params?: ListParams): Promise<User[]> {
    const res = await this.http.request<{ users: User[] }>("GET", "/v1/server/users", {
      auth: "apiKey",
      query: listQuery(params),
    });
    return res.users ?? [];
  }

  async get(id: string): Promise<User> {
    return this.http.request<User>("GET", `/v1/server/users/${id}`, { auth: "apiKey" });
  }

  async update(id: string, input: { status?: string }): Promise<User> {
    return this.http.request<User>("PATCH", `/v1/server/users/${id}`, {
      auth: "apiKey",
      body: input,
    });
  }

  async delete(id: string): Promise<void> {
    await this.http.request<void>("DELETE", `/v1/server/users/${id}`, { auth: "apiKey" });
  }
}
