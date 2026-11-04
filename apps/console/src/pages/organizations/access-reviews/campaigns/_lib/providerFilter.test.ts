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

import { isSourceExcluded, providerFilterOptions } from "./providerFilter";

describe("providerFilterOptions", () => {
  it("lists each provider once, labeled for display, and skips sources with none", () => {
    expect(providerFilterOptions([
      { provider: "SLACK", displayName: "Slack" },
      { provider: "SLACK", displayName: "Slack" },
      { provider: "GITHUB", displayName: "GitHub" },
      { provider: null, displayName: null },
      { provider: "OVHCLOUD", displayName: "OVHcloud" },
      { provider: "OVHCLOUD_US", displayName: "OVHcloud" },
    ])).toEqual([
      { value: "GITHUB", label: "GitHub" },
      { value: "OVHCLOUD", label: "OVHcloud, EU" },
      { value: "OVHCLOUD_US", label: "OVHcloud, US" },
      { value: "SLACK", label: "Slack" },
    ]);
  });
});

describe("isSourceExcluded", () => {
  const slack = { provider: "SLACK" };
  const github = { provider: "GITHUB" };
  const csv = { provider: null };

  it("keeps every source when no provider is selected", () => {
    const filters = { providers: [] };
    expect(isSourceExcluded(filters, slack)).toBe(false);
    expect(isSourceExcluded(filters, csv)).toBe(false);
  });

  it("keeps only the selected providers", () => {
    const filters = { providers: ["SLACK", "GITHUB"] };
    expect(isSourceExcluded(filters, slack)).toBe(false);
    expect(isSourceExcluded(filters, github)).toBe(false);
    expect(isSourceExcluded({ providers: ["SLACK"] }, github)).toBe(true);
  });

  it("hides sources with no provider while a provider is selected", () => {
    expect(isSourceExcluded({ providers: ["SLACK"] }, csv)).toBe(true);
  });
});
