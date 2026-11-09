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

import { safeContinueURL } from "./useSafeContinueURL";

const origin = "https://eu.oborp.com";

describe("safeContinueURL", () => {
  it("falls back when a same-origin continue path is scheme-relative", () => {
    const url = safeContinueURL("https://eu.oborp.com//www.wikipedia.org", origin);

    expect(url.origin).toBe(origin);
    expect(url.href).toBe(`${origin}/`);
  });

  it("uses the caller fallback when that continue path is scheme-relative", () => {
    const url = safeContinueURL(
      "https://eu.oborp.com//www.wikipedia.org",
      origin,
      "/organizations/abc",
    );

    expect(url.href).toBe(`${origin}/organizations/abc`);
  });

  it("keeps a same-origin path and search", () => {
    const url = safeContinueURL("/organizations/abc?x=1", origin);

    expect(url.href).toBe(`${origin}/organizations/abc?x=1`);
  });

  it("keeps an absolute same-origin path and search", () => {
    const url = safeContinueURL("https://eu.oborp.com/organizations/abc?x=1#hash", origin);

    expect(url.href).toBe(`${origin}/organizations/abc?x=1`);
  });

  it("keeps the path of a foreign absolute URL on this origin", () => {
    const url = safeContinueURL("https://evil.example/organizations/abc?x=1", origin);

    expect(url.origin).toBe(origin);
    expect(url.href).toBe(`${origin}/organizations/abc?x=1`);
  });

  it("uses the fallback when continue is missing", () => {
    const url = safeContinueURL(null, origin, "/organizations/abc");

    expect(url.href).toBe(`${origin}/organizations/abc`);
  });

  it("uses the page origin when continue and fallback are missing", () => {
    const url = safeContinueURL(null, origin);

    expect(url.href).toBe(`${origin}/`);
  });

  it("uses the fallback when continue cannot be parsed", () => {
    const url = safeContinueURL("https://[", origin, "/organizations/abc");

    expect(url.href).toBe(`${origin}/organizations/abc`);
  });

  it("ignores a scheme-relative fallback", () => {
    const url = safeContinueURL(null, origin, "//www.wikipedia.org");

    expect(url.origin).toBe(origin);
    expect(url.href).toBe(`${origin}/`);
  });
});
