/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_FLEET_ENDPOINT: string;
  readonly VITE_FLEET_PROJECT_ID: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
