// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import faroUploader from "@grafana/faro-rollup-plugin";
import type { Plugin, PluginOption } from "vite";

export const faroMetaBlock = [
  "{{if .FaroEnabled}}",
  "<meta name=\"faro-collector-url\" content=\"{{.FaroCollectorURL}}\">",
  "<meta name=\"faro-api-key\" content=\"{{.FaroAPIKey}}\">",
  "{{end}}",
].join("");

export const faroCollectorOriginTemplate
  = "{{if .FaroEnabled}} {{.FaroCollectorOrigin}}{{end}}";

export function omitFaroMeta(html: string): string {
  return html.replaceAll(faroMetaBlock, "");
}

export function omitFaroCollectorOrigin(policy: string): string {
  return policy.replaceAll(faroCollectorOriginTemplate, "");
}

// Drops the Faro Go actions while Vite is serving. The production build keeps
// them so probod can render the meta tags.
export function omitFaroDevPlaceholders(): Plugin {
  return {
    name: "omit-faro-dev-placeholders",
    apply: "serve",
    transformIndexHtml: {
      order: "pre",
      handler(html) {
        return omitFaroMeta(html);
      },
    },
  };
}

export interface FaroSourcemapEnv {
  FARO_SOURCEMAP_ENDPOINT?: string;
  FARO_SOURCEMAP_API_KEY?: string;
  FARO_STACK_ID?: string;
  GITHUB_SHA?: string;
}

export interface FaroSourcemapUploadOptions {
  appName: string;
  appId: string | undefined;
  prefixPath?: string;
}

export interface FaroSourcemapUpload {
  sourcemap: false | "hidden";
  plugins: PluginOption[];
}

const sourcemapsOff: FaroSourcemapUpload = {
  sourcemap: false,
  plugins: [],
};

function credential(value: string | undefined): string | null {
  const trimmed = value?.trim() ?? "";
  if (trimmed === "") {
    return null;
  }

  return trimmed;
}

// A blank credential skips the uploader. Local and staging builds have none.
export function faroSourcemapUpload(
  options: FaroSourcemapUploadOptions,
  env: FaroSourcemapEnv = process.env,
): FaroSourcemapUpload {
  const endpoint = credential(env.FARO_SOURCEMAP_ENDPOINT);
  const apiKey = credential(env.FARO_SOURCEMAP_API_KEY);
  const stackId = credential(env.FARO_STACK_ID);
  const appId = credential(options.appId);
  if (
    endpoint === null
    || apiKey === null
    || stackId === null
    || appId === null
  ) {
    return sourcemapsOff;
  }

  const gitHash = credential(env.GITHUB_SHA);

  return {
    sourcemap: "hidden",
    plugins: [
      faroUploader({
        appName: options.appName,
        appId,
        endpoint,
        apiKey,
        stackId,
        gzipContents: true,
        keepSourcemaps: false,
        recursive: true,
        verbose: true,
        ...(gitHash === null ? {} : { gitHash }),
        ...(options.prefixPath === undefined
          ? {}
          : { prefixPath: options.prefixPath }),
      }),
    ],
  };
}
