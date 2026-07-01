/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_GRAVITON_ENDPOINT: string;
  readonly VITE_GRAVITON_PROJECT_ID: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
