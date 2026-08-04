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
import {
  Button,
  Card,
  IconPlusLarge,
  PageHeader,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { MalaysiaPDPABreachesPage_organization$key } from "#/__generated__/core/MalaysiaPDPABreachesPage_organization.graphql";
import type { MalaysiaPDPABreachesPagePaginationQuery } from "#/__generated__/core/MalaysiaPDPABreachesPagePaginationQuery.graphql";
import type { MalaysiaPDPABreachesPageQuery } from "#/__generated__/core/MalaysiaPDPABreachesPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { MalaysiaPDPABreachListItem } from "./_components/MalaysiaPDPABreachListItem";

export const malaysiaPDPABreachesPageQuery = graphql`
  query MalaysiaPDPABreachesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...MalaysiaPDPABreachesPage_organization
      }
    }
  }
`;

const organizationFragment = graphql`
  fragment MalaysiaPDPABreachesPage_organization on Organization
  @refetchable(queryName: "MalaysiaPDPABreachesPagePaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
  ) {
    id
    canCreateMalaysiaPDPABreach: permission(
      action: "core:malaysia-pdpa-breach:create"
    )
    malaysiaPDPABreachIncidents(
      first: $first
      after: $after
      orderBy: { direction: DESC, field: AWARENESS_AT }
    ) @connection(
      key: "MalaysiaPDPABreachesPage__malaysiaPDPABreachIncidents"
      filters: []
    ) {
      totalCount
      edges {
        node {
          id
          ...MalaysiaPDPABreachListItem_incident
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

interface MalaysiaPDPABreachesPageProps {
  queryRef: PreloadedQuery<MalaysiaPDPABreachesPageQuery>;
}

export default function MalaysiaPDPABreachesPage({
  queryRef,
}: MalaysiaPDPABreachesPageProps) {
  const { t } = useTranslation("organizations/data-breaches");
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<MalaysiaPDPABreachesPageQuery>(
    malaysiaPDPABreachesPageQuery,
    queryRef,
  );

  usePageTitle(t("list.title"));

  if (data.organization?.__typename !== "Organization") {
    throw new Error("PAGE_NOT_FOUND: organization not found");
  }

  const pagination = usePaginationFragment<
    MalaysiaPDPABreachesPagePaginationQuery,
    MalaysiaPDPABreachesPage_organization$key
  >(organizationFragment, data.organization);
  const incidents = pagination.data.malaysiaPDPABreachIncidents.edges.map(
    ({ node }) => node,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("list.title")}
        description={t("list.description", {
          count: pagination.data.malaysiaPDPABreachIncidents.totalCount,
        })}
      >
        {pagination.data.canCreateMalaysiaPDPABreach && (
          <Button
            icon={IconPlusLarge}
            to={`/organizations/${organizationId}/data-breaches/new`}
          >
            {t("list.actions.create")}
          </Button>
        )}
      </PageHeader>

      {incidents.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("list.columns.incident")}</Th>
                    <Th>{t("list.columns.status")}</Th>
                    <Th>{t("list.columns.awarenessAt")}</Th>
                    <Th>{t("list.columns.affectedSubjects")}</Th>
                    <Th>{t("list.columns.recommendation")}</Th>
                    <Th>{t("list.columns.commissionerDeadline")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {incidents.map(incident => (
                    <MalaysiaPDPABreachListItem
                      key={incident.id}
                      incidentKey={incident}
                    />
                  ))}
                </Tbody>
              </Table>

              {pagination.hasNext && (
                <div className="border-t border-border-low p-4">
                  <Button
                    variant="secondary"
                    disabled={pagination.isLoadingNext}
                    onClick={() => pagination.loadNext(20)}
                  >
                    {pagination.isLoadingNext
                      ? t("list.actions.loading")
                      : t("list.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="space-y-3 py-12 text-center">
                <h2 className="text-lg font-semibold">{t("list.empty.title")}</h2>
                <p className="text-sm text-txt-tertiary">
                  {t("list.empty.description")}
                </p>
                {pagination.data.canCreateMalaysiaPDPABreach && (
                  <div className="flex justify-center pt-1">
                    <Button
                      icon={IconPlusLarge}
                      to={`/organizations/${organizationId}/data-breaches/new`}
                    >
                      {t("list.actions.create")}
                    </Button>
                  </div>
                )}
              </div>
            </Card>
          )}
    </div>
  );
}
