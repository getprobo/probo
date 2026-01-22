import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({
      exclude: ["src/pages/iam/**/*"],
      babel: {
        plugins: [
          [
            "relay",
            {
              eagerEsModules: true,
              artifactDirectory: "src/__generated__/core",
            },
          ],
        ],
      },
    }),
    react({
      include: ["src/pages/iam/**/*"],
      babel: {
        plugins: [
          [
            "relay",
            {
              eagerEsModules: true,
              artifactDirectory: "src/__generated__/iam",
            },
          ],
        ],
      },
    }),
    tailwindcss(),
  ],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "/environments": fileURLToPath(
        new URL("./src/environments", import.meta.url),
      ),
      "/type": fileURLToPath(new URL("./src/type.ts", import.meta.url)),
      "/components": fileURLToPath(
        new URL("./src/components", import.meta.url),
      ),
      "/coredata": fileURLToPath(new URL("./src/coredata", import.meta.url)),
      "/hooks": fileURLToPath(new URL("./src/hooks", import.meta.url)),
      "/layouts": fileURLToPath(new URL("./src/layouts", import.meta.url)),
      "/pages": fileURLToPath(new URL("./src/pages", import.meta.url)),
      "/routes": fileURLToPath(new URL("./src/routes", import.meta.url)),
      "/providers": fileURLToPath(new URL("./src/providers", import.meta.url)),
      "/permissions": fileURLToPath(
        new URL("./src/permissions", import.meta.url),
      ),
    },
  },
});
