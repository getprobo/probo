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

import { getWebInstrumentations, initializeFaro } from "@grafana/faro-web-sdk";
import { TracingInstrumentation } from "@grafana/faro-web-tracing";

export interface InitFaroOptions {
  appName: string;
  collectorURL: string;
  apiKey: string;
  apiURL?: string;
}

export function initFaro(options: InitFaroOptions): void {
  initializeFaro({
    url: options.collectorURL,
    apiKey: options.apiKey,
    app: {
      name: options.appName,
    },
    instrumentations: [
      ...getWebInstrumentations({ captureConsole: false }),
      new TracingInstrumentation({
        instrumentationOptions: {
          propagateTraceHeaderCorsUrls: traceHeaderCorsUrls(options.apiURL),
        },
      }),
    ],
  });
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
