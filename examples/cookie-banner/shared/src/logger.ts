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

import log from "loglevel";

// persist=false: loglevel writes `localStorage[key] = level` (not
// setItem). Those assignments happen at module load, before
// <probo-cookie-banner> connects and StorageDetector wraps Storage, so
// a persisted setLevel would only ever be seen as PRE_EXISTING.
log.setLevel("debug", false);

const namedLoggers = new Map<string, log.Logger>();

export function getExampleLogger(suffix: string): log.Logger {
  const name = `probo:example:${suffix}`;
  const existing = namedLoggers.get(name);
  if (existing) {
    return existing;
  }

  const logger = log.getLogger(name);
  logger.setLevel("debug", false);
  namedLoggers.set(name, logger);
  return logger;
}

export function enableNamedLoggers(): void {
  const keys = ["loglevel"];

  for (const [name, logger] of namedLoggers) {
    logger.setLevel("debug", false);
    keys.push(`loglevel:${name}`);
  }

  // Replay through setItem after the StorageDetector wrap is live.
  // Do not wait for probo-ready: discovery never fires that event
  // (`CookieBannerClient.loaded` is false without a published config).
  // Catch per key so one quota failure does not skip the rest.
  for (const key of keys) {
    try {
      localStorage.setItem(key, "DEBUG");
    } catch {
      // sandboxed frames and quota errors must not abort banner startup
    }
  }
}

export function scheduleEnableNamedLoggers(): void {
  queueMicrotask(() => {
    enableNamedLoggers();
  });
}
