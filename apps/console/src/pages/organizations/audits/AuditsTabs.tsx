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

import type { AuditsTabsQuery } from "#/__generated__/core/AuditsTabsQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const auditsTabsQuery = graphql`
  query AuditsTabsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        canListAudits: permission(action: "core:audit:list")
        canListAuditPrograms: permission(action: "core:audit-program:list")
      }
    }
  }
`;

interface AuditsTabsProps {
  queryRef: PreloadedQuery<AuditsTabsQuery>;
}

export function AuditsTabs({ queryRef }: AuditsTabsProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<AuditsTabsQuery>(auditsTabsQuery, queryRef);

  const canListAudits = data.organization?.canListAudits ?? false;
  const canListAuditPrograms
    = data.organization?.canListAuditPrograms ?? false;
  const baseUrl = `/organizations/${organizationId}/audits`;

  return (
    <Tabs>
      {canListAudits && (
        <TabLink to={baseUrl} end>
          {t("auditsTabs.audits")}
        </TabLink>
      )}
      {canListAuditPrograms && (
        <TabLink to={`${baseUrl}/programs`} end>
          {t("auditsTabs.programs")}
        </TabLink>
      )}
    </Tabs>
  );
}
