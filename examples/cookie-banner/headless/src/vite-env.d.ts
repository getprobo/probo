/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly PUBLIC_COOKIE_BANNER_ID?: string;
  readonly PUBLIC_COOKIE_BANNER_API_BASE_URL?: string;
  readonly PUBLIC_EXAMPLE_DOMAIN?: string;
  readonly PUBLIC_THEMED_BANNER_ID?: string;
  readonly PUBLIC_THEMED_TCF_BANNER_ID?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
