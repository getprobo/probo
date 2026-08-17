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

import { createDefaultIntegrations, resolveGoogleConsentMode } from "./index";

describe("createDefaultIntegrations", () => {
  it("includes Google Consent Mode by default", () => {
    expect(createDefaultIntegrations()).toHaveLength(1);
    expect(createDefaultIntegrations({})).toHaveLength(1);
    expect(createDefaultIntegrations({ googleConsentMode: true })).toHaveLength(1);
  });

  it("is empty when Google Consent Mode is disabled", () => {
    expect(createDefaultIntegrations({ googleConsentMode: false })).toHaveLength(0);
  });
});

describe("resolveGoogleConsentMode", () => {
  it("keeps the integration when the attribute is absent", () => {
    expect(resolveGoogleConsentMode(null)).toBe(true);
  });

  it("disables on \"off\" and \"false\"", () => {
    expect(resolveGoogleConsentMode("off")).toBe(false);
    expect(resolveGoogleConsentMode("false")).toBe(false);
  });

  it("normalizes case and whitespace", () => {
    expect(resolveGoogleConsentMode("OFF")).toBe(false);
    expect(resolveGoogleConsentMode("False")).toBe(false);
    expect(resolveGoogleConsentMode(" off ")).toBe(false);
  });

  it("keeps the integration for any other value", () => {
    expect(resolveGoogleConsentMode("on")).toBe(true);
    expect(resolveGoogleConsentMode("true")).toBe(true);
    expect(resolveGoogleConsentMode("")).toBe(true);
    expect(resolveGoogleConsentMode("nonsense")).toBe(true);
  });
});
