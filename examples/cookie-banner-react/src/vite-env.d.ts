/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly PUBLIC_COOKIE_BANNER_ID?: string;
  readonly PUBLIC_COOKIE_BANNER_API_BASE_URL?: string;
  readonly PUBLIC_POSTHOG_API_KEY?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
