import { Fleet, FleetError } from "@fleet/sdk";

function env(name: string, fallback?: string): string {
  const value = process.env[name] ?? fallback;
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

function log(title: string, data: unknown) {
  console.log(`\n=== ${title} ===`);
  console.log(typeof data === "string" ? data : JSON.stringify(data, null, 2));
}

function suffix() {
  return Date.now().toString(36);
}

export async function runServerDemo(fleet: Fleet) {
  log("Server: health", await fleet.server.health.check());

  const projects = await fleet.server.projects.list();
  log("Server: list projects", projects.map((p) => ({ id: p.id, name: p.name })));

  const dbId = `sdk_demo_${suffix()}`;
  const collId = "posts";
  await fleet.server.databases.createDatabase({ id: dbId, name: `SDK Demo DB ${suffix()}` });
  await fleet.server.databases.createCollection(dbId, {
    id: collId,
    name: "Posts",
  });
  await fleet.server.databases.createAttribute(dbId, collId, {
    key: "title",
    type: "string",
    size: 256,
  });
  await fleet.server.databases.createAttribute(dbId, collId, {
    key: "views",
    type: "integer",
  });

  const serverDoc = await fleet.server.databases.createDocument(dbId, collId, {
    data: { title: "Hello from Server SDK", views: 1 },
  });
  log("Server: create document", serverDoc);

  const team = await fleet.server.teams.create({ name: `SDK Team ${suffix()}` });
  log("Server: create team", { id: team.id, name: team.name, total: team.total });

  const users = await fleet.server.users.list({ page_size: 5 });
  log("Server: list users (top 5)", users.map((u) => ({ id: u.id, email: u.email })));

  return { dbId, collId, teamId: team.id };
}

export async function runClientDemo(
  fleet: Fleet,
  input: { dbId: string; collId: string; inviteEmail: string }
) {
  const email = env("FLEET_DEMO_EMAIL", `sdk.demo.${suffix()}@fleet.local`);
  const password = env("FLEET_DEMO_PASSWORD", "Sdk@123456");

  let signIn;
  try {
    signIn = await fleet.account.signIn({ email, password });
    log("Client: sign in", { id: signIn.account.id, email: signIn.account.email });
  } catch (err) {
    if (err instanceof FleetError && (err.status === 401 || err.status === 404)) {
      signIn = await fleet.account.signUp({
        email,
        password,
        name: "SDK Demo User",
      });
      log("Client: sign up", { id: signIn.account.id, email: signIn.account.email });
    } else {
      throw err;
    }
  }

  const me = await fleet.account.me();
  log("Client: me", { id: me.id, email: me.email, name: me.name });

  await fleet.account.updatePrefs({ theme: "sdk-demo", lang: "zh" });
  log("Client: prefs", await fleet.account.getPrefs());

  const clientDoc = await fleet.databases.createDocument(input.dbId, input.collId, {
    data: { title: "Hello from Client SDK", views: 2 },
  });
  log("Client: create document", clientDoc);

  const docs = await fleet.databases.listDocuments(input.dbId, input.collId, { page_size: 10 });
  log("Client: list documents", docs);

  const team = await fleet.teams.createTeam(`Client Team ${suffix()}`);
  log("Client: create team", team);

  await fleet.account.refresh(signIn.tokens.refresh_token);
  log("Client: refresh token (pick up team roles)", { ok: true });

  const invite = await fleet.teams.createMembership(team.id, {
    email: input.inviteEmail,
    name: "Invited Member",
    roles: ["member"],
  });
  log("Client: invite member (pending)", {
    id: invite.id,
    email: invite.email,
    status: invite.status,
  });

  const teams = await fleet.teams.listTeams();
  log("Client: list teams", teams.map((t) => ({ id: t.id, name: t.name, total: t.total })));
}

export function readEnv(name: string, fallback?: string): string {
  return env(name, fallback);
}
