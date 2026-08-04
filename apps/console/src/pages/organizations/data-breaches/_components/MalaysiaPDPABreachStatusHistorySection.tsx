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

import {
  Button,
  Card,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  usePaginationFragment,
} from "react-relay";

import type { MalaysiaPDPABreachStatusHistorySection_incident$key } from "#/__generated__/core/MalaysiaPDPABreachStatusHistorySection_incident.graphql";
import type { MalaysiaPDPABreachStatusHistorySectionPaginationQuery } from "#/__generated__/core/MalaysiaPDPABreachStatusHistorySectionPaginationQuery.graphql";

import { MalaysiaPDPABreachStatusHistoryListItem } from "./MalaysiaPDPABreachStatusHistoryListItem";

const incidentFragment = graphql`
  fragment MalaysiaPDPABreachStatusHistorySection_incident on MalaysiaPDPABreachIncident
  @refetchable(queryName: "MalaysiaPDPABreachStatusHistorySectionPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
  ) {
    id
    statusHistory(
      first: $first
      after: $after
      orderBy: { direction: DESC, field: CREATED_AT }
    ) @connection(
      key: "MalaysiaPDPABreachStatusHistorySection__statusHistory"
      filters: []
    ) {
      totalCount
      edges {
        node {
          id
          ...MalaysiaPDPABreachStatusHistoryListItem_history
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

interface MalaysiaPDPABreachStatusHistorySectionProps {
  incidentKey: MalaysiaPDPABreachStatusHistorySection_incident$key;
}

export function MalaysiaPDPABreachStatusHistorySection({
  incidentKey,
}: MalaysiaPDPABreachStatusHistorySectionProps) {
  const { t } = useTranslation("organizations/data-breaches");
  const pagination = usePaginationFragment<
    MalaysiaPDPABreachStatusHistorySectionPaginationQuery,
    MalaysiaPDPABreachStatusHistorySection_incident$key
  >(incidentFragment, incidentKey);
  const history = pagination.data.statusHistory.edges.map(({ node }) => node);

  return (
    <Card>
      <div className="space-y-1 border-b border-border-low p-6">
        <h2 className="text-base font-medium">{t("history.title")}</h2>
        <p className="text-sm text-txt-tertiary">
          {t("history.description", {
            count: pagination.data.statusHistory.totalCount,
          })}
        </p>
      </div>
      <Table>
        <Thead>
          <Tr>
            <Th>{t("history.columns.transition")}</Th>
            <Th>{t("history.columns.reason")}</Th>
            <Th>{t("history.columns.changedAt")}</Th>
            <Th>{t("history.columns.changedBy")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {history.map(entry => (
            <MalaysiaPDPABreachStatusHistoryListItem
              key={entry.id}
              historyKey={entry}
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
  );
}
