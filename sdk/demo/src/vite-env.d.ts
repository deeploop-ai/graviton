/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ORIONID_ENDPOINT: string;
  readonly VITE_ORIONID_PROJECT_ID: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
