/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly POSTHOG_KEY: string;
  readonly POSTHOG_HOST: string;
  readonly API_SERVER_HOST: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
