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

import { describe, expect, it } from "vitest";

import { rewriteLegacyConsoleHost } from "./hosted-url";

describe("rewriteLegacyConsoleHost", () => {
  it("rewrites the EU console host", () => {
    const url = new URL("https://eu.console.getprobo.com/api/cookie-banner/v1/");
    expect(rewriteLegacyConsoleHost(url).href).toBe(
      "https://eu.probo.com/api/cookie-banner/v1/",
    );
  });

  it("rewrites the US console host", () => {
    const url = new URL("https://us.console.getprobo.com/api/cookie-banner/v1/");
    expect(rewriteLegacyConsoleHost(url).href).toBe(
      "https://us.probo.com/api/cookie-banner/v1/",
    );
  });

  it("matches the host case-insensitively", () => {
    const url = new URL("https://EU.CONSOLE.GETPROBO.COM/api/cookie-banner/v1/");
    expect(rewriteLegacyConsoleHost(url).href).toBe(
      "https://eu.probo.com/api/cookie-banner/v1/",
    );
  });

  it("preserves path and query", () => {
    const url = new URL(
      "https://us.console.getprobo.com/api/cookie-banner/v1/banner/consents?lang=fr",
    );
    expect(rewriteLegacyConsoleHost(url).href).toBe(
      "https://us.probo.com/api/cookie-banner/v1/banner/consents?lang=fr",
    );
  });

  it("leaves already-migrated hosted URLs unchanged", () => {
    const url = new URL("https://eu.probo.com/api/cookie-banner/v1/");
    expect(rewriteLegacyConsoleHost(url)).toBe(url);
  });

  it("leaves self-hosted URLs unchanged", () => {
    const url = new URL("https://privacy.example.com/api/cookie-banner/v1/");
    expect(rewriteLegacyConsoleHost(url)).toBe(url);
  });

  it("does not rewrite suffix hosts", () => {
    const url = new URL("https://foo.eu.console.getprobo.com/api/cookie-banner/v1/");
    expect(rewriteLegacyConsoleHost(url)).toBe(url);
  });
});
