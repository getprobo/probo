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
import type { BannerConfig, TCFGVL, TCFRuntime } from "@probo/cookie-banner";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { grantForAction, startTCF } from "./api";
import { encodeTCString } from "./encode";
import { setLastTCString } from "./session";

const { update, runtimeHolder, cmpApiCtor } = vi.hoisted(() => ({
  update: vi.fn(),
  runtimeHolder: { current: null as TCFRuntime | null },
  cmpApiCtor: vi.fn(),
}));

vi.mock("@iabtechlabtcf/cmpapi", () => ({
  CmpApi: class {
    constructor(cmpId: number, cmpVersion: number, gdprApplies: boolean) {
      cmpApiCtor(cmpId, cmpVersion, gdprApplies);
    }

    update = update;
  },
}));

vi.mock("@probo/cookie-banner", () => ({
  setLayoutRenderer: vi.fn(),
  setTCFRuntime: (next: TCFRuntime) => {
    runtimeHolder.current = next;
  },
  getTCFRuntime: () => runtimeHolder.current,
}));

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
  },
  specialPurposes: {},
  features: {},
  specialFeatures: {},
  stacks: {},
  vendors: {
    "52": {
      id: 52,
      name: "Test Vendor",
      purposes: [1],
      legIntPurposes: [],
      flexiblePurposes: [],
      specialPurposes: [],
      features: [],
      specialFeatures: [],
      policyUrl: "https://example.com/privacy",
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
    },
    categories: [],
    texts: {},
    ...overrides,
  };
}

describe("startTCF displayStatus", () => {
  beforeEach(() => {
    update.mockReset();
    cmpApiCtor.mockReset();
    setLastTCString(undefined);
    startTCF();
  });

  afterEach(() => {
    setLastTCString(undefined);
  });

  it("reopens with the stored tc and uiVisible true", () => {
    const runtime = runtimeHolder.current;
    const stored = encodeTCString(bannerConfig(), "all");

    runtime?.onConfig(bannerConfig(), stored);
    const published = update.mock.calls[0]?.[0];
    update.mockClear();

    runtime?.onUIVisible?.(true);
    expect(update).toHaveBeenCalledWith(published, true);
    expect(TCString.decode(published as string).vendorConsents.has(52)).toBe(true);
  });

  it("hides with the stored tc and uiVisible false", () => {
    const runtime = runtimeHolder.current;
    const stored = encodeTCString(bannerConfig(), "all");

    runtime?.onConfig(bannerConfig(), stored);
    const published = update.mock.calls[0]?.[0];
    update.mockClear();

    runtime?.onUIVisible?.(false);
    expect(update).toHaveBeenCalledWith(published, false);
  });

  it("publishes a current-GVL TC before the visitor consents", () => {
    const runtime = runtimeHolder.current;

    runtime?.onConfig(bannerConfig());

    expect(update).toHaveBeenCalledTimes(1);
    const [tc, visible] = update.mock.calls[0] ?? [];
    expect(visible).toBe(true);
    expect(TCString.decode(tc as string).vendorListVersion).toBe(42);
    expect(TCString.decode(tc as string).vendorConsents.has(52)).toBe(false);
  });

  it("re-encodes stored choices onto the current GVL", () => {
    const runtime = runtimeHolder.current;
    const current = bannerConfig();
    const stored = encodeTCString(
      bannerConfig({
        tcf: {
          ...current.tcf!,
          gvl: { ...tcfGvl, vendorListVersion: 41 },
        },
      }),
      "all",
    );

    runtime?.onConfig(current, stored);

    expect(update).toHaveBeenCalledTimes(1);
    const [tc, visible] = update.mock.calls[0] ?? [];
    expect(visible).toBe(false);
    expect(tc).not.toBe(stored);
    expect(TCString.decode(tc as string).vendorListVersion).toBe(42);
    expect(TCString.decode(tc as string).vendorConsents.has(52)).toBe(true);
  });

  it("does not restore a stored TC when the TCF policy version changed", () => {
    const runtime = runtimeHolder.current;
    const current = bannerConfig();
    const stored = encodeTCString(
      bannerConfig({
        tcf: {
          ...current.tcf!,
          gvl: { ...tcfGvl, tcfPolicyVersion: 4 },
        },
      }),
      "all",
    );

    runtime?.onConfig(current, stored);

    expect(update).toHaveBeenCalledTimes(1);
    const [tc, visible] = update.mock.calls[0] ?? [];
    expect(visible).toBe(true);
    expect(TCString.decode(tc as string).vendorListVersion).toBe(42);
    expect(TCString.decode(tc as string).vendorConsents.has(52)).toBe(false);
  });

  it("does not construct CmpApi when TCF is inactive", () => {
    const runtime = runtimeHolder.current;

    runtime?.onConfig(bannerConfig({ regulation: "CCPA", tcf: {} }));
    update.mockClear();

    runtime?.onUIVisible?.(true);
    expect(update).not.toHaveBeenCalled();
    expect(cmpApiCtor).not.toHaveBeenCalled();
  });

  it("falls back to a current-GVL TC when CmpApi rejects the restored string", () => {
    const runtime = runtimeHolder.current;
    const stored = encodeTCString(bannerConfig(), "all");
    update.mockImplementation(() => {
      throw new Error("invalid tc");
    });

    runtime?.onConfig(bannerConfig(), stored);

    const last = update.mock.calls[update.mock.calls.length - 1];
    expect(last?.[1]).toBe(true);
    expect(TCString.decode(last?.[0] as string).vendorListVersion).toBe(42);
    expect(TCString.decode(last?.[0] as string).vendorConsents.has(52)).toBe(false);
  });

  it("maps ACKNOWLEDGE to the all-granted TC state", () => {
    expect(grantForAction("ACKNOWLEDGE", undefined)).toBe("all");
    expect(grantForAction("ACCEPT_ALL", undefined)).toBe("all");
    expect(grantForAction("REJECT_ALL", undefined)).toBe("none");
    expect(grantForAction("REJECT_ALL", {
      purposeConsents: [1],
      purposeLegitimateInterests: [],
      vendorConsents: [52],
      vendorLegitimateInterests: [],
      specialFeatureOptins: [],
    })).toBe("none");
  });

  it("constructs CmpApi with the instance cmp_id", () => {
    const runtime = runtimeHolder.current;

    runtime?.onConfig(bannerConfig({
      tcf: {
        gvl: tcfGvl,
        cmp_id: 123,
        cmp_version: 1,
        publisher_cc: "AA",
      },
    }));

    expect(cmpApiCtor).toHaveBeenCalledWith(123, 1, true);
  });

  it("throws when TCF is active without cmp_id", () => {
    const runtime = runtimeHolder.current;

    expect(() => runtime?.onConfig(bannerConfig({ tcf: { gvl: tcfGvl } }))).toThrow(
      /tcf.cmp_id is missing or invalid/,
    );
  });

  it("throws when TCF is active without cmp_version", () => {
    const runtime = runtimeHolder.current;

    expect(() =>
      runtime?.onConfig(bannerConfig({ tcf: { gvl: tcfGvl, cmp_id: 4095 } })),
    ).toThrow(/tcf.cmp_version is missing or invalid/);
  });

  it("previews withdrawn vendor consent on CmpApi while the UI is open", () => {
    const runtime = runtimeHolder.current;
    const cfg = bannerConfig();
    runtime?.onConfig(cfg);
    runtime?.onConsent?.("ACCEPT_ALL", cfg);
    runtime?.onUIVisible?.(true);
    update.mockClear();

    runtime?.setPendingChoices?.({
      purposeConsents: [1],
      purposeLegitimateInterests: [],
      vendorConsents: [],
      vendorLegitimateInterests: [],
      specialFeatureOptins: [],
    });

    expect(update).toHaveBeenCalledTimes(1);
    expect(update.mock.calls[0]?.[1]).toBe(true);
    const decoded = TCString.decode(update.mock.calls[0]?.[0] as string);
    expect(decoded.vendorConsents.has(52)).toBe(false);
  });

  it("persists a new TC string when a vendor consent is withdrawn", () => {
    const runtime = runtimeHolder.current;
    const cfg = bannerConfig();
    runtime?.onConfig(cfg);
    const accepted = runtime?.onConsent?.("ACCEPT_ALL", cfg);
    expect(TCString.decode(accepted!).vendorConsents.has(52)).toBe(true);

    runtime?.setPendingChoices?.({
      purposeConsents: [1],
      purposeLegitimateInterests: [],
      vendorConsents: [],
      vendorLegitimateInterests: [],
      specialFeatureOptins: [],
    });
    const customized = runtime?.onConsent?.("CUSTOMIZE", cfg);

    expect(customized).not.toBe(accepted);
    expect(TCString.decode(customized!).vendorConsents.has(52)).toBe(false);
  });

  it("persists a reject TC string after accept-all", () => {
    const runtime = runtimeHolder.current;
    const cfg = bannerConfig();
    runtime?.onConfig(cfg);
    const accepted = runtime?.onConsent?.("ACCEPT_ALL", cfg);
    expect(TCString.decode(accepted!).vendorConsents.has(52)).toBe(true);

    runtime?.onUIVisible?.(true);
    runtime?.setPendingChoices?.({
      purposeConsents: [1],
      purposeLegitimateInterests: [],
      vendorConsents: [52],
      vendorLegitimateInterests: [],
      specialFeatureOptins: [],
    });
    const rejected = runtime?.onConsent?.("REJECT_ALL", cfg);

    expect(rejected).not.toBe(accepted);
    expect(TCString.decode(rejected!).vendorConsents.has(52)).toBe(false);
  });

  it("returns the encoded TC string even if CmpApi.update throws", () => {
    const runtime = runtimeHolder.current;
    const cfg = bannerConfig();
    runtime?.onConfig(cfg);
    runtime?.onConsent?.("ACCEPT_ALL", cfg);
    update.mockImplementation(() => {
      throw new Error("cmpapi");
    });

    const rejected = runtime?.onConsent?.("REJECT_ALL", cfg);
    expect(rejected).toEqual(expect.any(String));
    expect(TCString.decode(rejected!).vendorConsents.has(52)).toBe(false);
  });
});
