import { fileURLToPath, URL } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({
      exclude: [
        "src/pages/iam/**/*",
        "src/components/connectors/**/*",
      ],
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
      include: [
        "src/pages/iam/**/*",
        "src/components/connectors/**/*",
      ],
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
      "#": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
