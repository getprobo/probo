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

import type { DataRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { FaroClient } from "./client";

const sdk = vi.hoisted(() => ({
  imports: 0,
  initFaro: vi.fn(),
}));

vi.mock("./sdk", () => {
  sdk.imports += 1;
  return { initFaro: sdk.initFaro };
});

const router = {} as DataRouter;

const collectorTags = {
  "faro-collector-url": "https://collector.example/collect",
  "faro-api-key": "secret",
};

function stubMetaTags(tags: Record<string, string>): void {
  vi.stubGlobal("document", {
    querySelector(selector: string) {
      const name = selector.match(/name="([^"]+)"/)?.[1];
      if (name === undefined || tags[name] === undefined) {
        return null;
      }
      return { getAttribute: () => tags[name] };
    },
  });
}

describe("FaroClient", () => {
  beforeEach(() => {
    sdk.imports = 0;
    sdk.initFaro.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("skips the SDK import when either collector tag is missing", async () => {
    const cases: Record<string, string>[] = [
      { "faro-api-key": collectorTags["faro-api-key"] },
      { "faro-collector-url": collectorTags["faro-collector-url"] },
    ];
    for (const tags of cases) {
      sdk.imports = 0;
      stubMetaTags(tags);
      const client = new FaroClient({ appName: "console" });

      await expect(client.start(router)).resolves.toBeUndefined();
      expect(client.enabled()).toBe(false);
      expect(sdk.imports).toBe(0);
    }
  });

  it("resolves without importing the SDK when document is unavailable", async () => {
    const client = new FaroClient({ appName: "console" });

    await expect(client.start(router)).resolves.toBeUndefined();
    expect(client.enabled()).toBe(false);
    expect(sdk.imports).toBe(0);
  });

  it("shares one SDK load across repeated start calls", async () => {
    stubMetaTags(collectorTags);
    const client = new FaroClient({
      appName: "console",
      apiURL: "https://api.example",
    });

    const first = client.start(router);
    const second = client.start(router);

    expect(first).toBe(second);
    await first;
    expect(sdk.imports).toBe(1);
    expect(sdk.initFaro).toHaveBeenCalledTimes(1);
    expect(sdk.initFaro).toHaveBeenCalledWith({
      appName: "console",
      collectorURL: collectorTags["faro-collector-url"],
      apiKey: collectorTags["faro-api-key"],
      apiURL: "https://api.example",
      router,
    });
  });
});
