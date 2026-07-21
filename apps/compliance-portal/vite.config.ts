// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { fileURLToPath, URL } from "node:url";

import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv, type Plugin } from "vite";

// index.html is also a Go html/template for the production SPA shell. Vite's
// dev server does not execute those actions, so bare {{if}}/{{range}} text is
// moved into <body> by the browser. Substitute safe defaults while serving.
function goHtmlTemplateDevDefaults(): Plugin {
  return {
    name: "go-html-template-dev-defaults",
    // Only register during `vite` / `vite serve` — production builds must keep
    // the Go html/template actions for pkg/server/trust.
    apply: "serve",
    transformIndexHtml: {
      order: "pre",
      handler(html) {
        return html
          .replace(
            /\{\{if \.HTMLLang\}\}\{\{\.HTMLLang\}\}\{\{else\}\}en\{\{end\}\}/g,
            "en",
          )
          .replace(
            /\{\{if \.FaviconURL\}\}\{\{\.FaviconURL\}\}\{\{else\}\}(\/favicons\/favicon\.ico)\{\{end\}\}/g,
            "$1",
          )
          .replace(
            /\{\{if \.CanonicalURL\}\}<link rel="canonical" href="\{\{\.CanonicalURL\}\}">\{\{end\}\}\s*/g,
            "",
          )
          .replace(
            // eslint-disable-next-line @stylistic/max-len
            /\{\{range \.Hreflang\}\}<link rel="alternate" hreflang="\{\{\.Lang\}\}" href="\{\{\.Href\}\}">\s*\{\{end\}\}\s*/g,
            "",
          )
          .replace(/\{\{\.Title\}\}/g, "Compliance")
          .replace(/\{\{\.Description\}\}/g, "")
          .replace(/\{\{\.OGURL\}\}/g, "");
      },
    },
  };
}

// https://vite.dev/config/
// @vitejs/plugin-react@6 (Vite 8) no longer runs Babel, so the Relay tagged
// template transform is applied via @rolldown/plugin-babel instead.
export default defineConfig(({ mode, command }) => {
  const envDir = fileURLToPath(new URL(".", import.meta.url));
  // Empty prefix: load non-VITE_ vars too (proxy target is Node-only).
  const env = loadEnv(mode, envDir, "");
  const proxyTarget = env.COMPLIANCE_PORTAL_PROXY_TARGET;

  if (command === "serve" && !proxyTarget) {
    throw new Error(
      "COMPLIANCE_PORTAL_PROXY_TARGET is required in apps/compliance-portal/.env",
    );
  }

  return {
    plugins: [
      goHtmlTemplateDevDefaults(),
      react(),
      babel({ plugins: ["relay"] }),
      tailwindcss(),
    ],
    build: {
      assetsDir: "assets",
      rolldownOptions: {
        output: {
          codeSplitting: {
            groups: [
              {
                name: "react",
                test: /node_modules\/(?:react-dom|react)\//,
              },
              {
                name: "relay",
                test: /node_modules\/(?:react-relay|relay-runtime)\//,
              },
              {
                name: "react-router",
                test: /node_modules\/react-router\//,
              },
            ],
          },
        },
      },
    },
    base: "./",
    server: {
      port: 5174,
      proxy: proxyTarget
        ? {
            "^/graphql": {
              // Host-routed compliance-portal API (trust-center HTTPS listener).
              target: proxyTarget,
              changeOrigin: true,
              secure: false, // local step-ca
            },
          }
        : undefined,
    },
    resolve: {
      alias: {
        "#": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },
  };
});
