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

export type ThirdPartySectionId
  = | "profile"
    | "services"
    | "supply-chain"
    | "stakeholders"
    | "assurance"
    | "agreements"
    | "risks"
    | "measures";

export type ThirdPartySectionGroup = "compliance" | "risk";

export interface ThirdPartySection {
  id: ThirdPartySectionId;
  path: string;
  labelKey: string;
  group?: ThirdPartySectionGroup;
}

export const THIRD_PARTY_SECTIONS: ThirdPartySection[] = [
  {
    id: "profile",
    path: "profile",
    labelKey: "nav.thirdPartyProfile",
  },
  {
    id: "services",
    path: "services",
    labelKey: "nav.thirdPartyServices",
  },
  {
    id: "supply-chain",
    path: "supply-chain",
    labelKey: "nav.thirdPartySupplyChain",
  },
  {
    id: "stakeholders",
    path: "stakeholders",
    labelKey: "nav.thirdPartyStakeholders",
    group: "compliance",
  },
  {
    id: "assurance",
    path: "assurance",
    labelKey: "nav.thirdPartyAssurance",
    group: "compliance",
  },
  {
    id: "agreements",
    path: "agreements",
    labelKey: "nav.thirdPartyAgreements",
    group: "compliance",
  },
  {
    id: "risks",
    path: "risks",
    labelKey: "nav.thirdPartyRisks",
    group: "risk",
  },
  {
    id: "measures",
    path: "measures",
    labelKey: "nav.thirdPartyMeasures",
    group: "risk",
  },
];

export const THIRD_PARTY_SECTION_GROUPS: {
  id: ThirdPartySectionGroup;
  labelKey: string;
}[] = [
  { id: "compliance", labelKey: "nav.thirdPartyCompliance" },
  { id: "risk", labelKey: "nav.thirdPartyRisk" },
];

export function firstThirdPartySection(): ThirdPartySection {
  return THIRD_PARTY_SECTIONS[0];
}

export function thirdPartyHref(
  organizationId: string,
  thirdPartyId: string,
  section: ThirdPartySection = firstThirdPartySection(),
): string {
  return `/organizations/${organizationId}/tprm/third-parties/${thirdPartyId}/${section.path}`;
}
