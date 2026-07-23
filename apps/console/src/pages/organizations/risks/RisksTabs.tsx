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

import { TabLink, Tabs } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { RisksTabsQuery } from "#/__generated__/core/RisksTabsQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const risksTabsQuery = graphql`
  query RisksTabsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        canListRisks: permission(action: "core:risk:list")
        canListRiskAssessments: permission(action: "core:risk-assessment:list")
      }
    }
  }
`;

interface RisksTabsProps {
  queryRef: PreloadedQuery<RisksTabsQuery>;
}

export function RisksTabs({ queryRef }: RisksTabsProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<RisksTabsQuery>(risksTabsQuery, queryRef);

  const canListRisks = data.organization?.canListRisks ?? false;
  const canListRiskAssessments
    = data.organization?.canListRiskAssessments ?? false;
  const baseUrl = `/organizations/${organizationId}`;

  return (
    <Tabs>
      {canListRisks && (
        <TabLink to={`${baseUrl}/risks`} end>
          {t("risksTabs.risks")}
        </TabLink>
      )}
      {canListRiskAssessments && (
        <TabLink to={`${baseUrl}/risk-assessments`} end>
          {t("risksTabs.riskAssessments")}
        </TabLink>
      )}
    </Tabs>
  );
}
