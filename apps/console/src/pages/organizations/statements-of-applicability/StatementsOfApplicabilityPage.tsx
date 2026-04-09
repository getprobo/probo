// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
import { useEffect } from "react";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type { StatementsOfApplicabilityPageFragment$key } from "#/__generated__/core/StatementsOfApplicabilityPageFragment.graphql";
import type { StatementsOfApplicabilityPagePaginationQuery } from "#/__generated__/core/StatementsOfApplicabilityPagePaginationQuery.graphql";
import type { StatementsOfApplicabilityPageQuery } from "#/__generated__/core/StatementsOfApplicabilityPageQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";

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
      filter: { type: "StatementOfApplicabilityFilter", defaultValue: { snapshotId: null } }
  ) {
      statementsOfApplicability(
          first: $first
          after: $after
          last: $last
          before: $before
          orderBy: $order
          filter: $filter
      ) @connection(key: "StatementsOfApplicabilityPage_statementsOfApplicability", filters: ["filter"]) {
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
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Statements of Applicability"));

  const { organization } = usePreloadedQuery(statementsOfApplicabilityPageQuery, queryRef);

  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { statementsOfApplicability },
    loadNext,
    hasNext,
    refetch,
    isLoadingNext,
  } = usePaginationFragment<
    StatementsOfApplicabilityPagePaginationQuery,
    StatementsOfApplicabilityPageFragment$key
  >(paginatedFragment, organization);

  useEffect(() => {
    if (snapshotId) {
      refetch(
        { filter: { snapshotId } },
        { fetchPolicy: "store-or-network" },
      );
    }
  }, [snapshotId, refetch]);

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Statements of Applicability")}
        description={__(
          "Manage statements of applicability for your organization's frameworks.",
        )}
      >
        {!isSnapshotMode
          && organization.canCreateStatementOfApplicability && (
          <CreateStatementOfApplicabilityDialog
            connectionId={statementsOfApplicability.__id}
          >
            <Button icon={IconPlusLarge}>
              {__("Add statement of applicability")}
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
                    <Th>{__("Name")}</Th>
                    <Th>{__("Created at")}</Th>
                    <Th>{__("Controls")}</Th>
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
                      ? __("Loading...")
                      : __("Load more")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No statements of applicability yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__(
                    "Create your first statement of applicability to get started.",
                  )}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
