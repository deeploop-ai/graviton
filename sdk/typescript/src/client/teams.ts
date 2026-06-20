import { listQuery, type HttpTransport } from "../http.js";
import type { ListParams, Membership, Team } from "../types.js";

export class ClientTeamsService {
  constructor(private readonly http: HttpTransport) {}

  async createTeam(name: string): Promise<Team> {
    return this.http.request<Team>("POST", "/v1/teams", { body: { name } });
  }

  async listTeams(params?: ListParams): Promise<Team[]> {
    const res = await this.http.request<{ teams: Team[] }>("GET", "/v1/teams", {
      query: listQuery(params),
    });
    return res.teams ?? [];
  }

  async getTeam(id: string): Promise<Team> {
    return this.http.request<Team>("GET", `/v1/teams/${id}`);
  }

  async deleteTeam(id: string): Promise<void> {
    await this.http.request<void>("DELETE", `/v1/teams/${id}`);
  }

  async createMembership(
    teamId: string,
    input: { email: string; name?: string; roles?: string[] }
  ): Promise<Membership> {
    return this.http.request<Membership>("POST", `/v1/teams/${teamId}/memberships`, {
      body: {
        team_id: teamId,
        email: input.email,
        name: input.name ?? "",
        roles: input.roles,
      },
    });
  }

  async listMemberships(teamId: string): Promise<Membership[]> {
    const res = await this.http.request<{ memberships: Membership[] }>(
      "GET",
      `/v1/teams/${teamId}/memberships`
    );
    return res.memberships ?? [];
  }

  async updateMembershipStatus(
    teamId: string,
    membershipId: string,
    status: "accepted" | "rejected"
  ): Promise<Membership> {
    return this.http.request<Membership>(
      "PATCH",
      `/v1/teams/${teamId}/memberships/${membershipId}/status`,
      {
        body: {
          team_id: teamId,
          membership_id: membershipId,
          status,
        },
      }
    );
  }

  async deleteMembership(teamId: string, membershipId: string): Promise<void> {
    await this.http.request<void>(
      "DELETE",
      `/v1/teams/${teamId}/memberships/${membershipId}`
    );
  }
}
