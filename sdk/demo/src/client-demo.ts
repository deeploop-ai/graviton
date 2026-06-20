import { Fleet } from "@fleet/sdk";
import { readEnv, runClientDemo } from "./flows.js";

const fleet = Fleet.create({
  endpoint: readEnv("FLEET_ENDPOINT", "http://localhost:8088"),
  projectId: readEnv("FLEET_PROJECT_ID", "default"),
});

runClientDemo(fleet, {
  dbId: readEnv("FLEET_DEMO_DB_ID"),
  collId: readEnv("FLEET_DEMO_COLL_ID", "posts"),
  inviteEmail: readEnv("FLEET_DEMO_INVITE_EMAIL", "invitee@fleet.local"),
}).catch((err) => {
  console.error(err);
  process.exit(1);
});
