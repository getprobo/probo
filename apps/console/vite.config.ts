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
import { createRequire } from "node:module";
import { isIP } from "node:net";
import { fileURLToPath, URL } from "node:url";

import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

const require = createRequire(import.meta.url);

const cspTemplatePath = fileURLToPath(
  new URL("./content-security-policy.txt.tmpl", import.meta.url),
);

// Process-lifetime nonce for Vite serve: pairs with html.cspNonce so the
// React Fast Refresh inline preamble is allowed without unsafe-inline.
const viteDevCspNonce = randomBytes(16).toString("base64");

// Same policy as production (apps/console/csp.go), with a script nonce for
// Fast Refresh and ws:/wss: on connect-src for Vite HMR.
function consoleContentSecurityPolicy(
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

function proxyTo(target: string) {
  return {
    target,
    changeOrigin: true,
    ws: true,
  };
}

function appOriginFromEnv(env: Record<string, string>): string {
  const explicit = env.CONSOLE_APP_ORIGIN?.trim();
  if (explicit) {
    return explicit.replace(/\/$/, "");
  }

  const apiURL = env.VITE_API_URL?.trim() || "http://localhost:8080";
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

// Origin after private /api/files/v1/{id} 307s — mirrors
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

// @vitejs/plugin-react@6 (Vite 8) no longer runs Babel, so the Relay tagged
// template transform is applied via @rolldown/plugin-babel instead. The iam
// pages and the rest of the app compile against separate artifact directories.
const iamFiles = /src[/\\]pages[/\\]iam[/\\]/;

// https://vite.dev/config/
export default defineConfig(({ mode, command }) => {
  const envDir = fileURLToPath(new URL(".", import.meta.url));
  // Empty prefix: load non-VITE_ vars too (CSP app origin is Node-only).
  const env = loadEnv(mode, envDir, "");
  const appOrigin = appOriginFromEnv(env);
  const fileStorageOrigin = fileStorageOriginFromEnv(env);

  return {
    // Expose CSP peer origins to the client for Markdown img allowlisting.
    define: {
      "import.meta.env.VITE_APP_ORIGIN": JSON.stringify(appOrigin),
      "import.meta.env.VITE_FILE_STORAGE_ORIGIN": JSON.stringify(
        fileStorageOrigin,
      ),
    },
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
    // Dev-only: Vite stamps scripts (incl. React Fast Refresh preamble) with
    // this nonce; production Go CSP does not use nonces.
    html: command === "serve" ? { cspNonce: viteDevCspNonce } : undefined,
    server: {
      headers: {
        "Content-Security-Policy": consoleContentSecurityPolicy(
          appOrigin,
          fileStorageOrigin,
          viteDevCspNonce,
        ),
        "X-Frame-Options": "DENY",
        "X-Content-Type-Options": "nosniff",
        "Referrer-Policy": "no-referrer",
        "Permissions-Policy": "microphone=(), camera=(), geolocation=()",
      },
      proxy: {
        "/api": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
        // Employee-portal Vite on 5175. Same-origin /employee-portal hrefs
        // then load that SPA's /employee-portal/@vite and assets from this
        // origin. Inverse of employee-portal proxying /auth and /me here.
        "/employee-portal": proxyTo("http://localhost:5175"),
      },
    },
    resolve: {
      alias: {
        "#": fileURLToPath(new URL("./src", import.meta.url)),
        "mermaid": require.resolve("mermaid/dist/mermaid.esm.min.mjs"),
      },
    },
  };
});
