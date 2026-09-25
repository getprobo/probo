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

import { afterEach, describe, expect, it, vi } from "vitest";

import { createDefaultIntegrations, resolveGcmEnabled } from "./index";

describe("createDefaultIntegrations", () => {
  it("includes Google Consent Mode by default", () => {
    expect(createDefaultIntegrations()).toHaveLength(1);
    expect(createDefaultIntegrations([])).toHaveLength(1);
    expect(createDefaultIntegrations([{ name: "gcm", enabled: true }])).toHaveLength(1);
  });

  it("is empty when Google Consent Mode is disabled", () => {
    expect(createDefaultIntegrations([{ name: "gcm", enabled: false }])).toHaveLength(0);
  });
});

describe("resolveGcmEnabled", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("defaults to enabled when the attribute is absent", () => {
    expect(resolveGcmEnabled(null)).toBe(true);
  });

  it("parses \"true\" and \"false\"", () => {
    expect(resolveGcmEnabled("true")).toBe(true);
    expect(resolveGcmEnabled("false")).toBe(false);
  });

  it("normalizes case and whitespace", () => {
    expect(resolveGcmEnabled("FALSE")).toBe(false);
    expect(resolveGcmEnabled(" false ")).toBe(false);
    expect(resolveGcmEnabled("True")).toBe(true);
  });

  it("warns and falls back to enabled on invalid values", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    expect(resolveGcmEnabled("off")).toBe(true);
    expect(resolveGcmEnabled("")).toBe(true);
    expect(resolveGcmEnabled("nonsense")).toBe(true);
    expect(warn).toHaveBeenCalledTimes(3);
  });
});
