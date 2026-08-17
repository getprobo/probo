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

import type { ComponentType } from "react";

import type { NavPanel_organization$data } from "#/__generated__/iam/NavPanel_organization.graphql";
import type { NavGroup, NavGroupKey } from "#/pages/iam/organizations/_lib/navigation";

import { AccessReviewNavPanel } from "./AccessReviewNavPanel";
import { CompliancePortalNavPanel } from "./CompliancePortalNavPanel";
import { GovernanceNavPanel } from "./GovernanceNavPanel";
import { ItamNavPanel } from "./ItamNavPanel";
import { PrivacyNavPanel } from "./PrivacyNavPanel";
import { RegistriesNavPanel } from "./RegistriesNavPanel";
import { RiskManagementNavPanel } from "./RiskManagementNavPanel";
import { SettingsNavPanel } from "./SettingsNavPanel";
import { TprmNavPanel } from "./TprmNavPanel";

export interface NavPanelBodyProps {
  organizationKey: NavPanel_organization$data;
  group: NavGroup;
}

export const navPanels: Record<NavGroupKey, ComponentType<NavPanelBodyProps>> = {
  governance: GovernanceNavPanel,
  riskManagement: RiskManagementNavPanel,
  tprm: TprmNavPanel,
  privacy: PrivacyNavPanel,
  itam: ItamNavPanel,
  registries: RegistriesNavPanel,
  compliancePortal: CompliancePortalNavPanel,
  accessReview: AccessReviewNavPanel,
  settings: SettingsNavPanel,
};
