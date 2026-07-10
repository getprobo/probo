// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { TabLink, Tabs } from "@probo/ui";
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
  const { __ } = useTranslate();
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
          {__("Risks")}
        </TabLink>
      )}
      {canListRiskAssessments && (
        <TabLink to={`${baseUrl}/risk-assessments`} end>
          {__("Risk assessments")}
        </TabLink>
      )}
    </Tabs>
  );
}
