import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  // Use the PUBLIC_ prefix so the example accepts the same env var names
  // (PUBLIC_COOKIE_BANNER_ID, PUBLIC_COOKIE_BANNER_API_BASE_URL,
  // PUBLIC_POSTHOG_API_KEY) used by probo.com. Copy .env.example to .env
  // and fill in your values.
  envPrefix: "PUBLIC_",
  server: {
    port: 5180,
  },
});
