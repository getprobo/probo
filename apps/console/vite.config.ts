import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { fileURLToPath, URL } from "node:url";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react({ babel: { plugins: ["relay"] } }), tailwindcss()],
  server: {
    proxy: {
      "/connect": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/authz": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "/type": fileURLToPath(new URL("./src/type.ts", import.meta.url)),
      "/components": fileURLToPath(
        new URL("./src/components", import.meta.url),
      ),
      "/coredata": fileURLToPath(
        new URL("./src/coredata", import.meta.url),
      ),
      "/hooks": fileURLToPath(new URL("./src/hooks", import.meta.url)),
      "/layouts": fileURLToPath(new URL("./src/layouts", import.meta.url)),
      "/pages": fileURLToPath(new URL("./src/pages", import.meta.url)),
      "/routes": fileURLToPath(new URL("./src/routes", import.meta.url)),
      "/providers": fileURLToPath(new URL("./src/providers", import.meta.url)),
      "/permissions": fileURLToPath(
        new URL("./src/permissions", import.meta.url),
      ),
      "/locales": fileURLToPath(
        new URL("../../packages/i18n/locales", import.meta.url),
      ),
    },
  },
});
