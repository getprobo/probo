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

import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { faroSettings } from "./settings";
import { faroCollectorOriginTemplate, faroMetaBlock } from "./vite";

const apps = ["console", "employee-portal", "compliance-portal"];

describe("faroSettings", () => {
  it("returns nothing when the collector URL or the API key is missing", () => {
    expect(faroSettings("", "secret")).toBeNull();
    expect(faroSettings("https://collector.example/collect", "")).toBeNull();
    expect(faroSettings("  ", "secret")).toBeNull();
  });

  it("returns trimmed settings when both values are set", () => {
    expect(
      faroSettings("  https://collector.example/collect  ", "  secret  "),
    ).toEqual({
      url: "https://collector.example/collect",
      apiKey: "secret",
    });
  });
});

describe("Faro app snippets", () => {
  it("matches the meta block and the CSP clause in each app", () => {
    for (const app of apps) {
      const root = new URL(`../../../apps/${app}/`, import.meta.url);
      const html = readFileSync(new URL("index.html", root), "utf8");
      const policy = readFileSync(
        new URL("content-security-policy.txt.tmpl", root),
        "utf8",
      );
      expect(html).toContain(faroMetaBlock);
      expect(policy).toContain(faroCollectorOriginTemplate);
    }
  });
});

describe("Faro public entry", () => {
  it("does not import the SDK until startFaro loads it", () => {
    const index = readFileSync(new URL("./index.ts", import.meta.url), "utf8");
    const settings = readFileSync(new URL("./settings.ts", import.meta.url), "utf8");
    expect(index).not.toContain("@grafana/faro");
    expect(settings).not.toContain("@grafana/faro");
    expect(index).toContain("import(\"./sdk\")");
  });
});
