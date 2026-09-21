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

import type { BannerConfig, TCFGVL } from "@probo/cookie-banner";
import { describe, expect, it, vi } from "vitest";

vi.mock("@probo/cookie-banner", () => ({
  BRANDING: "BRANDING",
  CLOSE_ICON: "CLOSE",
  esc: (s: string) => s.replace(/</g, "&lt;"),
  floatingCard: (_position: string, _aria: unknown, inner: string) => inner,
  getTCFRuntime: () => null,
  interpolate: (template: string, vars: Record<string, string>) =>
    template.replace(/\{\{(\w+)\}\}/g, (_, key: string) => vars[key] ?? ""),
}));

import { renderTCFLayout } from "./layout";

const vendorId = 52;

const tcfGvl: TCFGVL = {
  gvlSpecificationVersion: 3,
  vendorListVersion: 42,
  tcfPolicyVersion: 5,
  lastUpdated: "2026-01-15T17:00:00Z",
  purposes: {
    "1": {
      id: 1,
      name: "Store and/or access information on a device",
      description: "Cookies, device or similar online identifiers.",
    },
    "7": {
      id: 7,
      name: "Measure advertising performance",
      description: "Advertising performance can be measured.",
      illustrations: ["How often an ad was shown can be measured."],
    },
  },
  specialPurposes: {
    "1": {
      id: 1,
      name: "Ensure security, prevent and detect fraud, and fix errors",
      description: "Your data can be used to protect against fraud.",
    },
  },
  features: {
    "1": {
      id: 1,
      name: "Match and combine data from other data sources",
      description: "Information from offline sources can be combined.",
    },
  },
  dataCategories: {
    "1": {
      id: 1,
      name: "IP addresses",
      description: "Your IP address can be used.",
    },
  },
  specialFeatures: {
    "1": {
      id: 1,
      name: "Use precise geolocation data",
      description: "Your precise geolocation data can be used.",
    },
  },
  stacks: {
    "1": {
      id: 1,
      name: "Advertising",
      description: "Advertising stack",
      purposes: [7],
      specialFeatures: [],
    },
  },
  vendors: {
    [String(vendorId)]: {
      id: vendorId,
      name: "Test Vendor",
      purposes: [1],
      legIntPurposes: [7],
      flexiblePurposes: [],
      specialPurposes: [1],
      features: [1],
      specialFeatures: [1],
      policyUrl: "https://example.com/privacy",
      usesCookies: true,
      cookieMaxAgeSeconds: 86400,
      cookieRefresh: false,
      usesNonCookieAccess: true,
      dataDeclaration: [1],
      urls: [
        {
          privacy: "https://example.com/privacy",
          legIntClaim: "https://example.com/li",
        },
      ],
    },
  },
};

function bannerConfig(overrides: Partial<BannerConfig> = {}): BannerConfig {
  return {
    banner_id: "banner",
    version: 1,
    language: "en",
    default_language: "en",
    cookie_policy_url: "https://example.com/cookies",
    consent_expiry_days: 180,
    consent_mode: "OPT_IN",
    regulation: "GDPR",
    layout: {
      presentation: "OPT_IN",
      initial_state: "banner",
      reopen_state: "panel",
      default_non_necessary_granted: false,
      buttons: {
        accept_all: true,
        reject_all: true,
        customize: true,
        save: true,
      },
      settings_link: "default",
    },
    show_branding: false,
    resource_reporting_enabled: false,
    tcf: {
      gvl: tcfGvl,
      cmp_id: 4095,
      cmp_version: 1,
      publisher_cc: "AA",
      policy_version: 5,
    },
    categories: [],
    texts: {},
    ...overrides,
  };
}

describe("renderTCFLayout", () => {
  it("returns null when TCF does not apply", () => {
    expect(renderTCFLayout(bannerConfig({ tcf: undefined }), "bottom-left")).toBeNull();
    expect(renderTCFLayout(bannerConfig({ regulation: "CCPA" }), "bottom-left")).toBeNull();
  });

  it("renders IAB first-layer disclosures and the second-layer lists", () => {
    const html = renderTCFLayout(bannerConfig(), "bottom-left");
    expect(html).toContain("Store and/or access information on a device");
    expect(html).toContain("Use precise geolocation data");
    expect(html).toContain("1 partner");
    expect(html).toContain("View partners");
    expect(html).toContain("Test Vendor");
    expect(html).toContain("probo_consent cookie for 180 days");
    expect(html).toContain("Advertising");
    expect(html).toContain('data-text="tcf_section_purposes"');
    expect(html).toContain('data-text="tcf_section_other_purposes"');
    expect(html).toContain('class="tcf-group"');
    expect(html).toContain('class="tcf-subsection"');
    expect(html).toContain('data-tcf="purpose-li"');
    expect(html).toContain('data-text="tcf_disclosure_store"');
    expect(html).toContain('data-text="tcf_panel_description"');
    expect(html).not.toContain('data-text="panel_description"');
    expect(html).toContain("How often an ad was shown can be measured.");
    expect(html).toContain("Ensure security, prevent and detect fraud, and fix errors");
    expect(html).toContain("Match and combine data from other data sources");
    expect(html).toContain("IP addresses");
    expect(html).toContain("Cookies (up to 1 day)");
    expect(html).toContain("Non-cookie storage");
    expect(html).toContain('href="https://example.com/privacy"');
    expect(html).toContain('href="https://example.com/li"');
    expect(html).toContain("Privacy policy");
    expect(html).not.toContain("Privacy policy: https://example.com/privacy");
    expect(html).not.toContain("probo-category-list");
  });

  it("widens the preference panel and distinguishes choice from disclosure rows", () => {
    const html = renderTCFLayout(bannerConfig(), "bottom-left");
    expect(html).not.toBeNull();
    if (html == null) {
      return;
    }
    expect(html).toContain('class="tcf-panel"');
    expect(html).toContain("max-width: 720px");
    expect(html).toContain('class="tcf-row-id" aria-hidden="true"></div>');
    expect(html).not.toContain('class="tcf-row-id" aria-hidden="true">1<');
    expect(html).not.toContain('class="tcf-row-id" aria-hidden="true">52<');
    expect(html).toContain('data-text="tcf_label_consent"');
    expect(html).toContain('data-text="tcf_label_li"');
    expect(html).toContain('data-text="tcf_label_optin"');
    expect(html).toContain('data-text="tcf_label_always_on"');
    expect(html).not.toContain('data-text="tcf_label_information"');
    expect(html).toContain('data-text="tcf_section_more"');
    expect(html).toContain("<details");
    expect(html).toContain('id="probo-tcf-vendors"');
    expect(html.indexOf('data-text="tcf_section_purposes"')).toBeLessThan(html.indexOf("Advertising"));
    expect(html.indexOf('data-tcf="special-feature"')).toBeLessThan(html.indexOf('id="probo-tcf-vendors"'));
    expect(html.indexOf('id="probo-tcf-vendors"')).toBeLessThan(
      html.indexOf('data-text="tcf_section_special_purposes"'),
    );

    const disclosureRows = html.split('class="tcf-row tcf-row-disclosure"').slice(1);
    expect(disclosureRows.length).toBeGreaterThan(0);
    for (const part of disclosureRows) {
      const row = part.split(/class="tcf-row tcf-row-(?:choice|disclosure)"/)[0];
      expect(row).not.toContain("<input");
    }
  });

  it("omits unused disclosure catalogs and rejects non-http policy URLs", () => {
    const html = renderTCFLayout(
      bannerConfig({
        tcf: {
          gvl: {
            ...tcfGvl,
            specialPurposes: {
              "2": {
                id: 2,
                name: "Deliver and present advertising and content",
                description: "Unused special purpose.",
              },
            },
            features: {
              "2": {
                id: 2,
                name: "Link different devices",
                description: "Unused feature.",
              },
            },
            dataCategories: {
              "2": {
                id: 2,
                name: "Device characteristics",
                description: "Unused category.",
              },
            },
            vendors: {
              [String(vendorId)]: {
                ...tcfGvl.vendors[String(vendorId)],
                specialPurposes: [],
                features: [],
                dataDeclaration: [],
                usesCookies: false,
                usesNonCookieAccess: false,
                policyUrl: "javascript:alert(1)",
                urls: [{ privacy: "javascript:alert(1)" }],
              },
            },
          },
          cmp_id: 4095,
          cmp_version: 1,
          publisher_cc: "AA",
          policy_version: 5,
        },
      }),
      "bottom-left",
    );

    expect(html).not.toContain("Deliver and present advertising and content");
    expect(html).not.toContain("Link different devices");
    expect(html).not.toContain("Device characteristics");
    expect(html).not.toContain("Cookies (up to");
    expect(html).not.toContain("javascript:");
    expect(html).not.toContain('href="javascript:');
  });
});
