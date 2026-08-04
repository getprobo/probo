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

import { usePageTitle } from "@probo/hooks";
import { Badge, Breadcrumb, PageHeader } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { MalaysiaPDPABreachDetailPageQuery } from "#/__generated__/core/MalaysiaPDPABreachDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { MalaysiaPDPABreachEditForm } from "./_components/MalaysiaPDPABreachEditForm";
import { MalaysiaPDPABreachStatusHistorySection } from "./_components/MalaysiaPDPABreachStatusHistorySection";
import { MalaysiaPDPABreachSummarySection } from "./_components/MalaysiaPDPABreachSummarySection";
import { MalaysiaPDPABreachTransitionForm } from "./_components/MalaysiaPDPABreachTransitionForm";
import { getBreachStatusBadgeVariant } from "./_lib/breachDisplay";

export const malaysiaPDPABreachDetailPageQuery = graphql`
  query MalaysiaPDPABreachDetailPageQuery($breachId: ID!) {
    incident: node(id: $breachId) {
      __typename
      ... on MalaysiaPDPABreachIncident {
        title
        description
        status
        ...MalaysiaPDPABreachSummarySection_incident
        ...MalaysiaPDPABreachTransitionForm_incident
        ...MalaysiaPDPABreachEditForm_incident
        ...MalaysiaPDPABreachStatusHistorySection_incident
      }
    }
  }
`;

interface MalaysiaPDPABreachDetailPageProps {
  queryRef: PreloadedQuery<MalaysiaPDPABreachDetailPageQuery>;
}

export default function MalaysiaPDPABreachDetailPage({
  queryRef,
}: MalaysiaPDPABreachDetailPageProps) {
  const { t } = useTranslation("organizations/data-breaches");
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<MalaysiaPDPABreachDetailPageQuery>(
    malaysiaPDPABreachDetailPageQuery,
    queryRef,
  );

  usePageTitle(
    data.incident?.__typename === "MalaysiaPDPABreachIncident"
      ? data.incident.title
      : t("detail.fallbackTitle"),
  );

  if (data.incident?.__typename !== "MalaysiaPDPABreachIncident") {
    throw new Error("PAGE_NOT_FOUND: breach incident not found");
  }

  const listUrl = `/organizations/${organizationId}/data-breaches`;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          { label: t("list.title"), to: listUrl },
          { label: data.incident.title },
        ]}
      />
      <PageHeader
        title={data.incident.title}
        description={data.incident.description || t("detail.description")}
      >
        <Badge variant={getBreachStatusBadgeVariant(data.incident.status)}>
          {t(`statuses.${data.incident.status}`)}
        </Badge>
      </PageHeader>

      <MalaysiaPDPABreachSummarySection incidentKey={data.incident} />
      <MalaysiaPDPABreachTransitionForm incidentKey={data.incident} />
      <MalaysiaPDPABreachEditForm incidentKey={data.incident} />
      <MalaysiaPDPABreachStatusHistorySection incidentKey={data.incident} />
    </div>
  );
}
