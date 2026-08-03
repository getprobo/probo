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

import { Button, Card, PageHeader, Table, Tbody, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { ThirdPartyFindingsPage_thirdParty$key } from "#/__generated__/core/ThirdPartyFindingsPage_thirdParty.graphql";
import type { ThirdPartyFindingsPageQuery } from "#/__generated__/core/ThirdPartyFindingsPageQuery.graphql";
import type { ThirdPartyFindingsPageRefetchQuery } from "#/__generated__/core/ThirdPartyFindingsPageRefetchQuery.graphql";

import { ThirdPartyFindingListItem } from "./_components/ThirdPartyFindingListItem";

const thirdPartyFindingsFragment = graphql`
  fragment ThirdPartyFindingsPage_thirdParty on ThirdParty
  @refetchable(queryName: "ThirdPartyFindingsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    findings(
      first: $first
      after: $after
      orderBy: { direction: DESC, field: CREATED_AT }
    ) @connection(key: "ThirdPartyFindingsPage_findings", filters: []) {
      edges {
        node {
          id
          ...ThirdPartyFindingListItem_finding
        }
      }
    }
  }
`;

export const thirdPartyFindingsPageQuery = graphql`
  query ThirdPartyFindingsPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        ...ThirdPartyFindingsPage_thirdParty
      }
    }
  }
`;

interface ThirdPartyFindingsPageProps {
  queryRef: PreloadedQuery<ThirdPartyFindingsPageQuery>;
}

export function ThirdPartyFindingsPage({ queryRef }: ThirdPartyFindingsPageProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<ThirdPartyFindingsPageQuery>(
    thirdPartyFindingsPageQuery,
    queryRef,
  );
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const { data: thirdParty, ...pagination } = usePaginationFragment<
    ThirdPartyFindingsPageRefetchQuery,
    ThirdPartyFindingsPage_thirdParty$key
  >(thirdPartyFindingsFragment, data.node);
  const findings = thirdParty.findings.edges.map(edge => edge.node);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartyFindingsPage.title")}
        description={t("thirdPartyFindingsPage.description")}
      />
      {findings.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("thirdPartyFindingsPage.columns.referenceId")}</Th>
                    <Th>{t("thirdPartyFindingsPage.columns.description")}</Th>
                    <Th>{t("thirdPartyFindingsPage.columns.status")}</Th>
                    <Th>{t("thirdPartyFindingsPage.columns.priority")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {findings.map(finding => (
                    <ThirdPartyFindingListItem
                      key={finding.id}
                      findingKey={finding}
                    />
                  ))}
                </Tbody>
              </Table>
              {pagination.hasNext && (
                <div className="border-t p-4">
                  <Button
                    variant="secondary"
                    disabled={pagination.isLoadingNext}
                    onClick={() => pagination.loadNext(50)}
                  >
                    {pagination.isLoadingNext
                      ? t("thirdPartyFindingsPage.loading")
                      : t("thirdPartyFindingsPage.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <p className="py-8 text-center text-txt-secondary">
                {t("thirdPartyFindingsPage.empty")}
              </p>
            </Card>
          )}
    </div>
  );
}
