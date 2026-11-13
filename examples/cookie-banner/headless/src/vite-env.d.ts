/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly PUBLIC_COOKIE_BANNER_ID?: string;
  readonly PUBLIC_COOKIE_BANNER_API_BASE_URL?: string;
  readonly PUBLIC_EXAMPLE_DOMAIN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
