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

import type { StatementsOfApplicabilityPageFragment$key } from "#/__generated__/core/StatementsOfApplicabilityPageFragment.graphql";
import type { StatementsOfApplicabilityPagePaginationQuery } from "#/__generated__/core/StatementsOfApplicabilityPagePaginationQuery.graphql";
import type { StatementsOfApplicabilityPageQuery } from "#/__generated__/core/StatementsOfApplicabilityPageQuery.graphql";

import { StatementOfApplicabilityRow } from "./_components/StatementOfApplicabilityRow";
import { CreateStatementOfApplicabilityDialog } from "./dialogs/CreateStatementOfApplicabilityDialog";

export const statementsOfApplicabilityPageQuery = graphql`
  query StatementsOfApplicabilityPageQuery($organizationId: ID!) {
      organization: node(id: $organizationId) {
          __typename
          ... on Organization {
              id
              canCreateStatementOfApplicability: permission(action: "core:statement-of-applicability:create")
              ...StatementsOfApplicabilityPageFragment
          }
      }
  }
`;

const paginatedFragment = graphql`
  fragment StatementsOfApplicabilityPageFragment on Organization
  @refetchable(queryName: "StatementsOfApplicabilityPagePaginationQuery")
  @argumentDefinitions(
      first: { type: "Int", defaultValue: 50 }
      order: {
          type: "StatementOfApplicabilityOrder"
          defaultValue: { direction: DESC, field: CREATED_AT }
      }
      after: { type: "CursorKey", defaultValue: null }
      before: { type: "CursorKey", defaultValue: null }
      last: { type: "Int", defaultValue: null }
  ) {
      statementsOfApplicability(
          first: $first
          after: $after
          last: $last
          before: $before
          orderBy: $order
      ) @connection(key: "StatementsOfApplicabilityPage_statementsOfApplicability") {
          __id
          edges {
              node {
                  id
                  ...StatementOfApplicabilityRowFragment
              }
          }
      }
  }
`;

export default function StatementsOfApplicabilityPage({
  queryRef,
}: {
  queryRef: PreloadedQuery<StatementsOfApplicabilityPageQuery>;
}) {
  const { t } = useTranslation();

  usePageTitle(t("statementsOfApplicabilityPage.title"));

  const { organization } = usePreloadedQuery<StatementsOfApplicabilityPageQuery>(
    statementsOfApplicabilityPageQuery,
    queryRef,
  );

  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { statementsOfApplicability },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    StatementsOfApplicabilityPagePaginationQuery,
    StatementsOfApplicabilityPageFragment$key
  >(paginatedFragment, organization);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("statementsOfApplicabilityPage.title")}
        description={t("statementsOfApplicabilityPage.description")}
      >
        {organization.canCreateStatementOfApplicability && (
          <CreateStatementOfApplicabilityDialog
            connectionId={statementsOfApplicability.__id}
          >
            <Button icon={IconPlusLarge}>
              {t("statementsOfApplicabilityPage.actions.add")}
            </Button>
          </CreateStatementOfApplicabilityDialog>
        )}
      </PageHeader>

      {statementsOfApplicability && statementsOfApplicability.edges.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("statementsOfApplicabilityPage.columns.name")}</Th>
                    <Th>{t("statementsOfApplicabilityPage.columns.createdAt")}</Th>
                    <Th>{t("statementsOfApplicabilityPage.columns.controls")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {statementsOfApplicability.edges.map(edge => (
                    <StatementOfApplicabilityRow
                      key={edge.node.id}
                      fKey={edge.node}
                      connectionId={statementsOfApplicability.__id}
                    />
                  ))}
                </Tbody>
              </Table>

              {hasNext && (
                <div className="p-4 border-t">
                  <Button
                    variant="secondary"
                    onClick={() => loadNext(50)}
                    disabled={isLoadingNext}
                  >
                    {isLoadingNext
                      ? t("statementsOfApplicabilityPage.actions.loading")
                      : t("statementsOfApplicabilityPage.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {t("statementsOfApplicabilityPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("statementsOfApplicabilityPage.empty.description")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
