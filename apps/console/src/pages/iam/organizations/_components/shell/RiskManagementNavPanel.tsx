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

import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { RiskManagementNavPanel_organization$key } from "#/__generated__/iam/RiskManagementNavPanel_organization.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";

import { NavPanelItem } from "./NavPanelItem";
import type { NavPanelBodyProps } from "./navPanels";

const riskManagementNavPanelFragment = graphql`
  fragment RiskManagementNavPanel_organization on Organization {
    canListRisks: permission(action: "core:risk:list")
    canListRiskAnalyses: permission(action: "core:risk-analysis:list")
  }
`;

export function RiskManagementNavPanel({ organizationKey, group }: NavPanelBodyProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment<RiskManagementNavPanel_organization$key>(
    riskManagementNavPanelFragment,
    organizationKey,
  );

  return (
    <>
      {organization.canListRisks && (
        <NavPanelItem
          label={t("nav.risks")}
          to={navHref(organizationId, group, "risks")}
        />
      )}
      {organization.canListRiskAnalyses && (
        <NavPanelItem
          label={t("nav.riskAnalyses")}
          to={navHref(organizationId, group, "risk-analyses")}
        />
      )}
    </>
  );
}
