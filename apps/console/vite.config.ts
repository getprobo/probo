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

// @vitejs/plugin-react@6 (Vite 8) no longer runs Babel, so the Relay tagged
// template transform is applied via @rolldown/plugin-babel instead. The iam
// pages and the rest of the app compile against separate artifact directories.
const iamFiles = /src[/\\]pages[/\\]iam[/\\]/;

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    babel({
      exclude: [/[/\\]node_modules[/\\]/, /\0rolldown[/\\]runtime\.js/, iamFiles],
      plugins: [
        [
          "relay",
          {
            eagerEsModules: true,
            artifactDirectory: "src/__generated__/core",
          },
        ],
      ],
    }),
    babel({
      include: /src[/\\]pages[/\\]iam[/\\].*\.[jt]sx?(?:$|\?)/,
      plugins: [
        [
          "relay",
          {
            eagerEsModules: true,
            artifactDirectory: "src/__generated__/iam",
          },
        ],
      ],
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
