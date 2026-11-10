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

import {
  createReactRouterV7DataOptions,
  ReactIntegration,
  withFaroRouterInstrumentation,
} from "@grafana/faro-react";
import { getWebInstrumentations, initializeFaro } from "@grafana/faro-web-sdk";
import { TracingInstrumentation } from "@grafana/faro-web-tracing";
import { type DataRouter, matchRoutes } from "react-router";

import { ignoredErrorPatterns } from "./ignore";
import { routePattern } from "./route";

export interface InitFaroOptions {
  appName: string;
  collectorURL: string;
  apiKey: string;
  apiURL?: string;
  router: DataRouter;
}

export function initFaro(options: InitFaroOptions): void {
  const { router } = options;
  initializeFaro({
    url: options.collectorURL,
    apiKey: options.apiKey,
    app: {
      name: options.appName,
    },
    ignoreErrors: ignoredErrorPatterns,
    pageTracking: {
      generatePageId: location => routePattern(
        router.routes,
        location.pathname,
        router.basename,
      ),
    },
    instrumentations: [
      ...getWebInstrumentations({ captureConsole: false }),
      new TracingInstrumentation({
        instrumentationOptions: {
          propagateTraceHeaderCorsUrls: traceHeaderCorsUrls(options.apiURL),
        },
      }),
      new ReactIntegration({
        router: createReactRouterV7DataOptions({ matchRoutes }),
      }),
    ],
  });

  // Faro types the v6 data router; React Router 8 exposes the same surface.
  withFaroRouterInstrumentation(
    router as unknown as Parameters<typeof withFaroRouterInstrumentation>[0],
  );
}

function traceHeaderCorsUrls(apiURL: string | undefined): RegExp[] {
  const origins = new Set<string>([window.location.origin]);
  const configured = apiURL?.trim();
  if (configured) {
    const formatted
      = configured.startsWith("http://") || configured.startsWith("https://")
        ? configured
        : `https://${configured}`;
    try {
      origins.add(new URL(formatted).origin);
    } catch {
      // A malformed API URL still leaves same-origin requests instrumented.
    }
  }

  return [...origins].map(origin => new RegExp(`^${escapeRegExp(origin)}`));
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
