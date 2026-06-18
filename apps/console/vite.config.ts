// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { createRequire } from "node:module";
import { fileURLToPath, URL } from "node:url";

import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const require = createRequire(import.meta.url);

// @vitejs/plugin-react v6 (rolldown) no longer runs Babel, so the Relay
// transform is applied through @rolldown/plugin-babel instead. The two
// instances mirror the previous core/iam split: every file outside
// src/pages/iam gets the "core" artifact directory, files inside it get "iam".
const iamPattern = /[/\\]src[/\\]pages[/\\]iam[/\\]/;

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    babel({
      plugins: [
        [
          "relay",
          { eagerEsModules: true, artifactDirectory: "src/__generated__/core" },
        ],
      ],
      exclude: [/[/\\]node_modules[/\\]/, iamPattern],
    }),
    babel({
      plugins: [
        [
          "relay",
          { eagerEsModules: true, artifactDirectory: "src/__generated__/iam" },
        ],
      ],
      include: [iamPattern],
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
      "mermaid": require.resolve("mermaid/dist/mermaid.esm.min.mjs"),
    },
  },
});
