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

import { TCString } from "@iabtechlabtcf/core";
import type { BannerConfig, TCFGVL } from "@probo/cookie-banner";
import { describe, expect, it } from "vitest";

import { encodeTCString, gdprApplies, requireCmpVersion } from "./encode";

const vendorId = 52;
const unusedVendorId = 99;

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
    "2": {
      id: 2,
      name: "Select basic ads",
      description: "Ads can be shown based on limited data.",
    },
  },
  specialPurposes: {},
  features: {},
  specialFeatures: {
    "1": {
      id: 1,
      name: "Use precise geolocation data",
      description: "Your precise geolocation data can be used.",
    },
  },
  stacks: {},
  vendors: {
    [String(vendorId)]: {
      id: vendorId,
      name: "Test Vendor",
      purposes: [1],
      legIntPurposes: [],
      flexiblePurposes: [],
      specialPurposes: [],
      features: [],
      specialFeatures: [1],
      policyUrl: "https://example.com/privacy",
      usesCookies: true,
      cookieMaxAgeSeconds: 86400,
      cookieRefresh: false,
      usesNonCookieAccess: false,
    },
    [String(unusedVendorId)]: {
      id: unusedVendorId,
      name: "Unused Vendor",
      purposes: [2],
      legIntPurposes: [],
      flexiblePurposes: [],
      specialPurposes: [],
      features: [],
      specialFeatures: [],
      policyUrl: "https://unused.example.com/privacy",
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
    consent_expiry_days: 365,
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

describe("gdprApplies", () => {
  it("is true for GDPR and UK_GDPR", () => {
    expect(gdprApplies(bannerConfig({ regulation: "GDPR" }))).toBe(true);
    expect(gdprApplies(bannerConfig({ regulation: "UK_GDPR" }))).toBe(true);
  });

  it("is false for other regulations", () => {
    expect(gdprApplies(bannerConfig({ regulation: "CCPA" }))).toBe(false);
    expect(gdprApplies(bannerConfig({ regulation: null }))).toBe(false);
  });
});

describe("encodeTCString", () => {
  it("throws when tcf.gvl is missing", () => {
    expect(() => encodeTCString(bannerConfig({ tcf: {} }), "all")).toThrow(
      /tcf.gvl is missing/,
    );
  });

  it("throws when tcf.cmp_id is missing", () => {
    expect(() => encodeTCString(bannerConfig({ tcf: { gvl: tcfGvl } }), "all")).toThrow(
      /tcf.cmp_id is missing or invalid/,
    );
  });

  it("throws when tcf.cmp_version is missing", () => {
    expect(() =>
      encodeTCString(bannerConfig({ tcf: { gvl: tcfGvl, cmp_id: 4095 } }), "all"),
    ).toThrow(/tcf.cmp_version is missing or invalid/);
  });

  it("throws when tcf.cmp_version exceeds the 12-bit field", () => {
    expect(() => requireCmpVersion(4096)).toThrow(/tcf.cmp_version is missing or invalid/);
  });

  it("encodes disclosed vendors on reject without granting consent", () => {
    const encoded = encodeTCString(bannerConfig(), "none");
    const decoded = TCString.decode(encoded);

    expect(decoded.cmpId).toBe(4095);
    expect(decoded.cmpVersion).toBe(1);
    expect(decoded.isServiceSpecific).toBe(true);
    expect(decoded.vendorsDisclosed.has(vendorId)).toBe(true);
    expect(decoded.vendorConsents.has(vendorId)).toBe(false);
    expect(decoded.purposeConsents.has(1)).toBe(false);
    expect(decoded.specialFeatureOptins.has(1)).toBe(false);
  });

  it("encodes vendor, purpose, and special-feature grants on accept", () => {
    const encoded = encodeTCString(bannerConfig(), "all");
    const decoded = TCString.decode(encoded);

    expect(decoded.vendorsDisclosed.has(vendorId)).toBe(true);
    expect(decoded.vendorConsents.has(vendorId)).toBe(true);
    expect(decoded.purposeConsents.has(1)).toBe(true);
    expect(decoded.specialFeatureOptins.has(1)).toBe(true);
  });

  it("encodes only the selected purpose, vendor, and special-feature bits", () => {
    const encoded = encodeTCString(bannerConfig(), {
      purposeConsents: [1],
      purposeLegitimateInterests: [],
      vendorConsents: [vendorId],
      vendorLegitimateInterests: [],
      specialFeatureOptins: [],
    });
    const decoded = TCString.decode(encoded);

    expect(decoded.purposeConsents.has(1)).toBe(true);
    expect(decoded.purposeConsents.has(2)).toBe(false);
    expect(decoded.purposeLegitimateInterests.has(1)).toBe(false);
    expect(decoded.vendorConsents.has(vendorId)).toBe(true);
    expect(decoded.vendorConsents.has(unusedVendorId)).toBe(false);
    expect(decoded.vendorLegitimateInterests.has(vendorId)).toBe(false);
    expect(decoded.specialFeatureOptins.has(1)).toBe(false);
  });
});
