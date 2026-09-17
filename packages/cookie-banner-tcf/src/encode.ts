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

import { GVL, TCModel, TCString, type VendorList } from "@iabtechlabtcf/core";
import type { BannerConfig, BannerTCF, TCFChoices } from "@probo/cookie-banner";

import { TCF_CMP_ID, TCF_CMP_VERSION } from "./constants";

export { TCF_CMP_ID, TCF_CMP_VERSION } from "./constants";

export function gdprApplies(config: BannerConfig): boolean {
  return config.regulation === "GDPR" || config.regulation === "UK_GDPR";
}

export type TCFGrant = "all" | "none" | TCFChoices;

export function encodeTCString(config: BannerConfig, grant: TCFGrant): string {
  const gvlJson = config.tcf?.gvl;
  if (!gvlJson) {
    throw new Error("cannot encode TC string: tcf.gvl is missing");
  }

  const tcf = config.tcf;
  const gvl = new GVL(vendorListForEncode(gvlJson));
  const tcModel = new TCModel(gvl);
  tcModel.cmpId = resolvedCmpID(tcf?.cmp_id);
  tcModel.cmpVersion = tcf?.cmp_version ?? TCF_CMP_VERSION;
  tcModel.isServiceSpecific = true;
  tcModel.consentScreen = 1;
  tcModel.publisherCountryCode = tcf?.publisher_cc ?? "AA";
  if (tcf?.policy_version) {
    tcModel.policyVersion = tcf.policy_version;
  }

  tcModel.setAllVendorsDisclosed();

  if (grant === "all") {
    tcModel.setAllVendorConsents();
    tcModel.setAllPurposeConsents();
    tcModel.setAllPurposeLegitimateInterests();
    tcModel.setAllVendorLegitimateInterests();
    tcModel.setAllSpecialFeatureOptins();
  } else if (grant !== "none") {
    applyChoices(tcModel, grant);
  }

  return TCString.encode(tcModel);
}

function applyChoices(tcModel: TCModel, choices: TCFChoices): void {
  tcModel.purposeConsents.set(choices.purposeConsents);
  tcModel.purposeLegitimateInterests.set(choices.purposeLegitimateInterests);
  tcModel.vendorConsents.set(choices.vendorConsents);
  tcModel.vendorLegitimateInterests.set(choices.vendorLegitimateInterests);
  tcModel.specialFeatureOptins.set(choices.specialFeatureOptins);
}

function resolvedCmpID(cmpId: number | undefined): number {
  if (cmpId !== undefined && cmpId > 1) {
    return cmpId;
  }

  return TCF_CMP_ID;
}

function vendorListForEncode(gvl: NonNullable<BannerTCF["gvl"]>): VendorList {
  return {
    gvlSpecificationVersion: gvl.gvlSpecificationVersion,
    vendorListVersion: gvl.vendorListVersion,
    tcfPolicyVersion: gvl.tcfPolicyVersion,
    lastUpdated: gvl.lastUpdated ?? new Date(0).toISOString(),
    purposes: (gvl.purposes ?? {}) as VendorList["purposes"],
    specialPurposes: (gvl.specialPurposes ?? {}) as VendorList["specialPurposes"],
    features: (gvl.features ?? {}) as VendorList["features"],
    specialFeatures: (gvl.specialFeatures ?? {}) as VendorList["specialFeatures"],
    stacks: (gvl.stacks ?? {}) as VendorList["stacks"],
    vendors: gvl.vendors as VendorList["vendors"],
  };
}
