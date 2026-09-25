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

import { projectConsentFromTCF } from "./project-consent";
import type { Category } from "./types";

function category(partial: Partial<Category> & Pick<Category, "slug" | "kind">): Category {
  return {
    name: partial.slug,
    description: "",
    cookies: [],
    gcm_consent_types: [],
    posthog_consent: false,
    tcf_purpose_ids: [],
    ...partial,
  };
}

describe("projectConsentFromTCF", () => {
  const categories = [
    category({ slug: "necessary", kind: "NECESSARY" }),
    category({ slug: "advertising", kind: "NORMAL", tcf_purpose_ids: [1, 3, 4] }),
    category({ slug: "analytics", kind: "NORMAL", tcf_purpose_ids: [1, 7, 8, 9, 10] }),
    category({ slug: "functional", kind: "NORMAL", tcf_purpose_ids: [] }),
  ];

  it("always grants necessary and fail-closes empty mappings", () => {
    expect(projectConsentFromTCF(categories, null)).toEqual({
      necessary: true,
      advertising: false,
      analytics: false,
      functional: false,
    });
  });

  it("requires every mapped purpose via consent, not legitimate interest", () => {
    expect(
      projectConsentFromTCF(categories, {
        purposeConsents: [1, 3],
        purposeLegitimateInterests: [4],
      }),
    ).toEqual({
      necessary: true,
      advertising: false,
      analytics: false,
      functional: false,
    });

    expect(
      projectConsentFromTCF(categories, {
        purposeConsents: [1, 3, 4],
        purposeLegitimateInterests: [],
      }),
    ).toEqual({
      necessary: true,
      advertising: true,
      analytics: false,
      functional: false,
    });
  });
});
