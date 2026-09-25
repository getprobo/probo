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

Object.assign(tcfGvl.vendors[String(vendorId)], {
  dataRetention: { stdRetention: 365, purposes: { "7": 30 } },
  deviceStorageDisclosureUrl: "https://example.com/storage",
});

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
    expect(
      renderTCFLayout(
        bannerConfig({
          layout: { ...bannerConfig().layout, presentation: "NOTICE" },
        }),
        "bottom-left",
      ),
    ).toBeNull();
  });

  it("renders IAB first-layer disclosures and the second-layer lists", () => {
    const html = renderTCFLayout(bannerConfig(), "bottom-left");
    expect(html).toContain("Store and/or access information on a device");
    expect(html).toContain("Use precise geolocation data");
    expect(html).toContain("1 partner");
    expect(html).toContain("View partners");
    const firstLayer = html!.split("<probo-preference-panel")[0];
    expect(firstLayer).toContain("unique identifiers and browsing data");
    expect(firstLayer).toContain('data-text="tcf_disclosure_data"');
    expect(firstLayer).toContain("service-specific");
    expect(firstLayer).not.toContain("service-specific consent");
    expect(firstLayer).toContain('data-text="tcf_disclosure_scope"');
    expect(firstLayer).toContain("withdraw or change your consent at any time");
    expect(firstLayer).toContain("Cookie settings");
    expect(firstLayer).toContain('data-text="tcf_disclosure_withdraw"');
    expect(firstLayer).toContain("legitimate interest");
    expect(firstLayer).toContain("object to that processing");
    expect(firstLayer).toContain('data-text="tcf_disclosure_object"');
    expect(firstLayer).toContain('class="tcf-disclosures"');
    expect(firstLayer).not.toContain('<p class="description" data-text="tcf_disclosure_store">');
    expect(firstLayer).not.toContain("How often an ad was shown can be measured.");
    expect(html).toContain("Test Vendor");
    expect(html).toContain("probo_consent cookie for 180 days");
    expect(html).toContain("Advertising");
    expect(html).toContain('data-text="tcf_section_purposes"');
    expect(html).toContain('data-text="tcf_section_other_purposes"');
    expect(html).toContain('class="tcf-group"');
    expect(html).toContain('class="tcf-subsection"');
    expect(html).toContain('data-tcf="purpose-li"');
    expect(html).toContain('class="tcf-purpose-vendors"');
    expect(html).toContain("1 partner seeking consent");
    expect(html).toContain("1 partner relying on legitimate interest");
    expect(html).not.toContain("0 partner");
    expect(html).toContain('data-text="tcf_disclosure_store"');
    expect(html).toContain('data-text="tcf_panel_description"');
    expect(html).not.toContain('data-text="panel_description"');
    const panel = html!.split("<probo-preference-panel")[1];
    expect(panel).toContain("unique identifiers and browsing data");
    expect(panel).toContain("service-specific");
    expect(panel).toContain('data-text="tcf_disclosure_data"');
    expect(panel).toContain('data-text="tcf_disclosure_scope"');
    expect(html).toContain("How often an ad was shown can be measured.");
    expect(html).toContain("Ensure security, prevent and detect fraud, and fix errors");
    expect(html).toContain("Match and combine data from other data sources");
    expect(html).toContain("IP addresses");
    expect(html).not.toContain('data-text="tcf_section_more"');
    expect(html).not.toContain("<details");
    expect(html).toContain("Cookies (up to 1 day)");
    expect(html).toContain("Non-cookie storage");
    expect(html).toContain('href="https://example.com/privacy"');
    expect(html).toContain('href="https://example.com/li"');
    expect(html).toContain("Privacy policy");
    expect(html).not.toContain("Privacy policy: https://example.com/privacy");
    expect(html).not.toContain("probo-category-list");
  });

  it("lists each vendor's purposes, legal bases, and other GVL declarations", () => {
    const html = renderTCFLayout(bannerConfig(), "bottom-left");
    const partners = html!.split('id="probo-tcf-vendors"')[1]?.split('data-text="tcf_section_storage"')[0] ?? "";
    expect(partners).toContain('class="tcf-vendor-details"');
    expect(partners).toContain('data-text="tcf_label_consent"');
    expect(partners).toContain("Store and/or access information on a device");
    expect(partners).toContain('data-text="tcf_label_li"');
    expect(partners).toContain("Measure advertising performance");
    expect(partners).toContain('data-text="tcf_section_special_purposes"');
    expect(partners).toContain("Ensure security, prevent and detect fraud, and fix errors");
    expect(partners).toContain('data-text="tcf_label_always_on"');
    expect(partners).toContain('data-text="tcf_section_features"');
    expect(partners).toContain("Match and combine data from other data sources");
    expect(partners).toContain('data-text="tcf_section_special_features"');
    expect(partners).toContain("Use precise geolocation data");
    expect(partners).toContain('data-text="tcf_section_data_categories"');
    expect(partners).toContain("IP addresses");
    expect(partners).toContain("Cookies (up to 1 day)");
    expect(partners).toContain("not refreshed");
    expect(partners).toContain("Purpose-specific storage");
    expect(partners).toContain('data-text="tcf_label_std_retention"');
    expect(partners).toContain("Standard retention");
    expect(partners).toContain("365 days");
    expect(partners).toContain("Measure advertising performance (30 days)");
    expect(partners).toContain('href="https://example.com/storage"');
    expect(partners).toContain("Device storage details");
    expect(partners).toContain('href="https://example.com/privacy"');
  });

  it("omits the legitimate-interest objection line when no vendor uses LI", () => {
    const vendor = tcfGvl.vendors[String(vendorId)];
    const html = renderTCFLayout(
      bannerConfig({
        tcf: {
          gvl: {
            ...tcfGvl,
            vendors: {
              [String(vendorId)]: { ...vendor, legIntPurposes: [] },
            },
          },
          cmp_id: 4095,
          cmp_version: 1,
          publisher_cc: "AA",
        },
      }),
      "bottom-left",
    );

    const [firstLayer, panel] = html!.split("<probo-preference-panel");
    expect(firstLayer).not.toContain('data-text="tcf_disclosure_object"');
    expect(firstLayer).not.toContain("object to that processing");
    expect(panel).not.toContain("unique identifiers and browsing data");
    expect(panel).not.toContain('data-text="tcf_disclosure_data"');
    expect(html).not.toContain('data-tcf="purpose-li"');
    expect(html).toContain("1 partner seeking consent");
    expect(html).not.toContain("relying on legitimate interest");
  });

  it("omits the vendor consent toggle when the vendor has no consent purposes", () => {
    const vendor = tcfGvl.vendors[String(vendorId)];
    const html = renderTCFLayout(
      bannerConfig({
        tcf: {
          gvl: {
            ...tcfGvl,
            vendors: {
              [String(vendorId)]: { ...vendor, purposes: [], flexiblePurposes: [], legIntPurposes: [7] },
            },
          },
          cmp_id: 4095,
          cmp_version: 1,
          publisher_cc: "AA",
        },
      }),
      "bottom-left",
    );

    const partners = html!.split('id="probo-tcf-vendors"')[1] ?? "";
    expect(partners).not.toContain('data-tcf="vendor-consent"');
    expect(partners).toContain('data-tcf="vendor-li"');
  });

  it("counts partners seeking consent or relying on LI for each purpose", () => {
    const extraVendor = {
      ...tcfGvl.vendors[String(vendorId)],
      id: 99,
      name: "Second Vendor",
      purposes: [1, 7],
      legIntPurposes: [],
    };
    const html = renderTCFLayout(
      bannerConfig({
        tcf: {
          gvl: {
            ...tcfGvl,
            vendors: {
              [String(vendorId)]: tcfGvl.vendors[String(vendorId)],
              "99": extraVendor,
            },
          },
          cmp_id: 4095,
          cmp_version: 1,
          publisher_cc: "AA",
        },
      }),
      "bottom-left",
    );

    const panel = html!.split("<probo-preference-panel")[1];
    const firstLayer = html!.split("<probo-preference-panel")[0];
    expect(firstLayer).not.toContain("seeking consent");
    expect(firstLayer).not.toContain("relying on legitimate interest");
    expect(panel).toContain("2 partners seeking consent");
    expect(panel).toContain("1 partner seeking consent · 1 partner relying on legitimate interest");
    expect(panel).toContain('class="tcf-purpose-vendors"');
    expect(panel.match(/class="tcf-purpose-vendors"/g)?.length).toBe(2);
  });

  it("widens the preference panel and orders purposes before partners", () => {
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
    expect(html).not.toContain('data-text="tcf_section_more"');
    expect(html).not.toContain("<details");
    expect(html).not.toContain("tcf-row-disclosure");
    expect(html).toContain('id="probo-tcf-vendors"');
    expect(html.indexOf('data-text="tcf_section_purposes"')).toBeLessThan(html.indexOf("Advertising"));
    expect(html.indexOf('data-tcf="special-feature"')).toBeLessThan(html.indexOf('id="probo-tcf-vendors"'));
    expect(html.indexOf('id="probo-tcf-vendors"')).toBeLessThan(
      html.indexOf('data-text="tcf_section_special_purposes"'),
    );
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
                dataRetention: undefined,
                policyUrl: "javascript:alert(1)",
                deviceStorageDisclosureUrl: "javascript:alert(1)",
                urls: [{ privacy: "javascript:alert(1)" }],
              } as TCFGVL["vendors"][string],
            },
          },
          cmp_id: 4095,
          cmp_version: 1,
          publisher_cc: "AA",
        },
      }),
      "bottom-left",
    );

    expect(html).not.toContain("Deliver and present advertising and content");
    expect(html).not.toContain("Link different devices");
    expect(html).not.toContain("Device characteristics");
    expect(html).not.toContain("Cookies (up to");
    expect(html).not.toContain("Standard retention");
    expect(html).not.toContain("Purpose-specific storage");
    expect(html).not.toContain("javascript:");
    expect(html).not.toContain('href="javascript:');
  });
});
