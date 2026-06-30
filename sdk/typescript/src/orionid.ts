import { AccountService, ClientDatabasesService, ClientTeamsService } from "./client/index.js";
import type { OrionidConfig } from "./http.js";
import { HttpTransport } from "./http.js";
import {
  APIKeysService,
  HealthService,
  ProjectsService,
  ServerDatabasesService,
  ServerTeamsService,
  StorageService,
  UsersService,
} from "./server/index.js";

export type { OrionidConfig } from "./http.js";
export { OrionidError } from "./errors.js";
export * from "./types.js";

export class Orionid {
  readonly account: AccountService;
  readonly databases: ClientDatabasesService;
  readonly teams: ClientTeamsService;

  readonly server: {
    health: HealthService;
    projects: ProjectsService;
    users: UsersService;
    teams: ServerTeamsService;
    databases: ServerDatabasesService;
    apiKeys: APIKeysService;
    storage: StorageService;
  };

  private readonly transport: HttpTransport;

  constructor(config: OrionidConfig) {
    this.transport = new HttpTransport(config);
    this.account = new AccountService(this.transport);
    this.databases = new ClientDatabasesService(this.transport);
    this.teams = new ClientTeamsService(this.transport);
    this.server = {
      health: new HealthService(this.transport),
      projects: new ProjectsService(this.transport),
      users: new UsersService(this.transport),
      teams: new ServerTeamsService(this.transport),
      databases: new ServerDatabasesService(this.transport),
      apiKeys: new APIKeysService(this.transport),
      storage: new StorageService(this.transport),
    };
  }

  static create(config: OrionidConfig): Orionid {
    return new Orionid(config);
  }

  /** Server API + optional Client API with a project API key. */
  static withApiKey(endpoint: string, projectId: string, apiKey: string): Orionid {
    return new Orionid({ endpoint, projectId, apiKey });
  }

  /** Client API with an existing user access token. */
  static withAccessToken(endpoint: string, projectId: string, accessToken: string): Orionid {
    return new Orionid({ endpoint, projectId, accessToken });
  }

  setAccessToken(token: string | undefined): void {
    this.transport.setAccessToken(token);
  }

  getAccessToken(): string | undefined {
    return this.transport.getAccessToken();
  }

  getProjectId(): string {
    return this.transport.getProjectId();
  }
}
