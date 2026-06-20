import { Fleet } from "@fleet/sdk";
import { readEnv, runServerDemo } from "./flows.js";

const fleet = Fleet.withApiKey(
  readEnv("FLEET_ENDPOINT", "http://localhost:8088"),
  readEnv("FLEET_PROJECT_ID", "default"),
  readEnv("FLEET_API_KEY")
);

runServerDemo(fleet).catch((err) => {
  console.error(err);
  process.exit(1);
});
