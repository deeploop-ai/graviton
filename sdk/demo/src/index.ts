import { Fleet, FleetError } from "@fleet/sdk";
import { readEnv, runClientDemo, runServerDemo } from "./flows.js";

async function main() {
  const endpoint = readEnv("FLEET_ENDPOINT", "http://localhost:8088");
  const projectId = readEnv("FLEET_PROJECT_ID", "default");
  const apiKey = readEnv("FLEET_API_KEY");

  console.log(`Fleet SDK demo → ${endpoint} (project: ${projectId})`);

  const serverFleet = Fleet.withApiKey(endpoint, projectId, apiKey);
  const { dbId, collId } = await runServerDemo(serverFleet);

  const clientFleet = Fleet.create({ endpoint, projectId });
  await runClientDemo(clientFleet, {
    dbId,
    collId,
    inviteEmail: readEnv("FLEET_DEMO_INVITE_EMAIL", "invitee@fleet.local"),
  });

  console.log("\nDemo completed successfully.");
}

main().catch((err) => {
  if (err instanceof FleetError) {
    console.error(`FleetError [${err.status}]: ${err.message}`);
  } else {
    console.error(err);
  }
  process.exit(1);
});
