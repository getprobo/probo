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

import { randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";
import { isIP } from "node:net";
import { fileURLToPath, URL } from "node:url";

import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv, type Plugin } from "vite";

const cspTemplatePath = fileURLToPath(
  new URL("./content-security-policy.txt.tmpl", import.meta.url),
);

// Process-lifetime nonce for Vite serve: pairs with html.cspNonce so the
// React Fast Refresh inline preamble is allowed without unsafe-inline.
const viteDevCspNonce = randomBytes(16).toString("base64");

// Same policy as production (apps/compliance-portal/csp.go), with a script
// nonce for Fast Refresh and ws:/wss: on connect-src for Vite HMR.
function compliancePortalContentSecurityPolicy(
  appOrigin: string,
  fileStorageOrigin: string,
  scriptNonce: string,
): string {
  const template = readFileSync(cspTemplatePath, "utf8");
  // Collapse newlines: Node rejects CR/LF in header values (Vite setHeader).
  return template
    .replaceAll("{{.AppOrigin}}", appOrigin)
    .replaceAll("{{.FileStorageOrigin}}", fileStorageOrigin)
    .trim()
    .replace(
      /script-src 'self';/,
      `script-src 'self' 'nonce-${scriptNonce}';`,
    )
    .replace(
      /connect-src 'self' ([^;]+);/,
      "connect-src 'self' $1 ws: wss:;",
    )
    .replace(/\s+/g, " ");
}

function appOriginFromEnv(env: Record<string, string>): string {
  const explicit = env.COMPLIANCE_PORTAL_APP_ORIGIN?.trim();
  if (explicit) {
    return explicit.replace(/\/$/, "");
  }

  const apiURL = env.VITE_API_URL?.trim();
  if (!apiURL) {
    return "";
  }

  const formatted
    = apiURL.startsWith("http://") || apiURL.startsWith("https://")
      ? apiURL
      : `https://${apiURL}`;

  return new URL(formatted).origin;
}

const defaultAWSRegion = "us-east-2";

function envBool(value: string | undefined, fallback: boolean): boolean {
  if (value == null || value.trim() === "") {
    return fallback;
  }

  switch (value.trim().toLowerCase()) {
    case "1":
    case "true":
    case "yes":
    case "on":
      return true;
    case "0":
    case "false":
    case "no":
    case "off":
      return false;
    default:
      return fallback;
  }
}

function validHostLabel(label: string): boolean {
  if (label.length === 0 || label.length > 63) {
    return false;
  }

  return /^[0-9A-Za-z-]+$/.test(label);
}

// Node's URL.hostname keeps IPv6 brackets ([::1]); isIP requires the bare form.
function hostnameIsIP(hostname: string): boolean {
  const bare
    = hostname.startsWith("[") && hostname.endsWith("]")
      ? hostname.slice(1, -1)
      : hostname;

  return isIP(bare) !== 0;
}

// Mirrors pkg/awsconfig.bucketIsVirtualHostable / aws-sdk-go-v2 rules.
function bucketIsVirtualHostable(bucket: string, https: boolean): boolean {
  if (isIP(bucket) !== 0) {
    return false;
  }

  const labels = https ? [bucket] : bucket.split(".");
  for (const label of labels) {
    if (label.length < 3 || label.length > 63) {
      return false;
    }

    if (/[A-Z]/.test(label)) {
      return false;
    }

    if (!validHostLabel(label)) {
      return false;
    }
  }

  return true;
}

// Origin after /api/files/v1/{id} 307s — mirrors
// pkg/awsconfig.CSPFileStorageOrigin from PROBOD_AWS_*.
function fileStorageOriginFromEnv(env: Record<string, string>): string {
  const endpoint = env.PROBOD_AWS_ENDPOINT?.trim() ?? "";
  const region = env.PROBOD_AWS_REGION?.trim() || defaultAWSRegion;
  const bucket = env.PROBOD_AWS_BUCKET?.trim() ?? "";
  const usePathStyle = envBool(env.PROBOD_AWS_USE_PATH_STYLE, false);

  if (!bucket) {
    return "";
  }

  if (!endpoint) {
    if (usePathStyle || !bucketIsVirtualHostable(bucket, true)) {
      return `https://s3.${region}.amazonaws.com`;
    }

    return `https://${bucket}.s3.${region}.amazonaws.com`;
  }

  const formatted
    = endpoint.startsWith("http://") || endpoint.startsWith("https://")
      ? endpoint
      : `https://${endpoint}`;
  const parsed = new URL(formatted);
  const https = parsed.protocol === "https:";

  if (
    usePathStyle
    || hostnameIsIP(parsed.hostname)
    || !bucketIsVirtualHostable(bucket, https)
  ) {
    return parsed.origin;
  }

  return `${parsed.protocol}//${bucket}.${parsed.host}`;
}

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
  const appOrigin = appOriginFromEnv(env);
  const fileStorageOrigin = fileStorageOriginFromEnv(env);

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
    // Dev-only: Vite stamps scripts (incl. React Fast Refresh preamble) with
    // this nonce; production Go CSP does not use nonces.
    html: command === "serve" ? { cspNonce: viteDevCspNonce } : undefined,
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
    // Absolute base: host-routed SPA lives at site root (`/:lang/...`). A
    // relative base (`./`) makes nested routes request `/en/documents/assets/`
    // instead of `/assets/`, so the SPA fallback serves HTML and browsers
    // block JS/CSS for the wrong MIME type.
    base: "/",
    server: {
      port: 5174,
      headers: {
        "Content-Security-Policy": compliancePortalContentSecurityPolicy(
          appOrigin,
          fileStorageOrigin,
          viteDevCspNonce,
        ),
        "X-Frame-Options": "DENY",
        "X-Content-Type-Options": "nosniff",
        "Referrer-Policy": "no-referrer",
        "Permissions-Policy": "microphone=(), camera=(), geolocation=()",
      },
      proxy: proxyTarget
        ? {
            // Host-routed compliance-portal API (trust-center HTTPS listener).
            // Match pathname only; req.url may include ?continue=… etc.
            "^/(?:graphql|initiate|callback|brand/(?:logo|dark-logo)|\\.well-known/oauth-client-metadata)(?:\\?|$)":
              {
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
