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

log.setLevel("debug");

const namedLoggers = new Map<string, log.Logger>();

export function getExampleLogger(suffix: string): log.Logger {
  const name = `probo:example:${suffix}`;
  const existing = namedLoggers.get(name);
  if (existing) {
    return existing;
  }

  const logger = log.getLogger(name);
  logger.setLevel("debug");
  namedLoggers.set(name, logger);
  return logger;
}

export function enableNamedLoggers(): void {
  for (const [name, logger] of namedLoggers) {
    logger.setLevel("debug");
    // loglevel persists via property assignment, which the SDK does not
    // hook. Replay the same key through setItem so a live write is
    // observed as SCRIPT after the detector has wrapped Storage.
    localStorage.setItem(`loglevel:${name}`, "DEBUG");
  }
}
