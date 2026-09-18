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

import type { BannerConfig, TCFGVL, TCFRuntime } from "@probo/cookie-banner";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { grantForAction, startTCF } from "./api";
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
  purposes: {},
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
      policy_version: 5,
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
    const stored = "C-stored-tc";

    runtime?.onConfig(bannerConfig(), stored);
    update.mockClear();

    runtime?.onUIVisible?.(true);
    expect(update).toHaveBeenCalledWith(stored, true);
  });

  it("hides with the stored tc and uiVisible false", () => {
    const runtime = runtimeHolder.current;
    const stored = "C-stored-tc";

    runtime?.onConfig(bannerConfig(), stored);
    update.mockClear();

    runtime?.onUIVisible?.(false);
    expect(update).toHaveBeenCalledWith(stored, false);
  });

  it("does not construct CmpApi when TCF is inactive", () => {
    const runtime = runtimeHolder.current;

    runtime?.onConfig(bannerConfig({ regulation: "CCPA", tcf: {} }));
    update.mockClear();

    runtime?.onUIVisible?.(true);
    expect(update).not.toHaveBeenCalled();
    expect(cmpApiCtor).not.toHaveBeenCalled();
  });

  it("falls back to an empty TC when the stored string is rejected", () => {
    const runtime = runtimeHolder.current;
    update.mockImplementationOnce(() => {
      throw new Error("invalid tc");
    });

    runtime?.onConfig(bannerConfig(), "C-bad-tc");

    expect(update).toHaveBeenCalledWith("", true);
  });

  it("maps ACKNOWLEDGE to the all-granted TC state", () => {
    expect(grantForAction("ACKNOWLEDGE", undefined)).toBe("all");
    expect(grantForAction("ACCEPT_ALL", undefined)).toBe("all");
    expect(grantForAction("REJECT_ALL", undefined)).toBe("none");
  });

  it("constructs CmpApi with the instance cmp_id", () => {
    const runtime = runtimeHolder.current;

    runtime?.onConfig(bannerConfig({
      tcf: {
        gvl: tcfGvl,
        cmp_id: 123,
        cmp_version: 1,
        publisher_cc: "AA",
        policy_version: 5,
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
});
