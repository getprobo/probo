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

import type { StatesOfApplicabilityPageFragment$key } from "#/__generated__/core/StatesOfApplicabilityPageFragment.graphql";
import type { StatesOfApplicabilityPagePaginationQuery } from "#/__generated__/core/StatesOfApplicabilityPagePaginationQuery.graphql";
import type { StatesOfApplicabilityPageQuery } from "#/__generated__/core/StatesOfApplicabilityPageQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";

import { StateOfApplicabilityRow } from "./_components/StateOfApplicabilityRow";
import { CreateStateOfApplicabilityDialog } from "./dialogs/CreateStateOfApplicabilityDialog";

export const statesOfApplicabilityPageQuery = graphql`
  query StatesOfApplicabilityPageQuery($organizationId: ID!) {
      organization: node(id: $organizationId) {
          __typename
          ... on Organization {
              id
              canCreateStateOfApplicability: permission(action: "core:state-of-applicability:create")
              ...StatesOfApplicabilityPageFragment
          }
      }
  }
`;

const paginatedFragment = graphql`
  fragment StatesOfApplicabilityPageFragment on Organization
  @refetchable(queryName: "StatesOfApplicabilityPagePaginationQuery")
  @argumentDefinitions(
      first: { type: "Int", defaultValue: 50 }
      order: {
          type: "StateOfApplicabilityOrder"
          defaultValue: { direction: DESC, field: CREATED_AT }
      }
      after: { type: "CursorKey", defaultValue: null }
      before: { type: "CursorKey", defaultValue: null }
      last: { type: "Int", defaultValue: null }
      filter: { type: "StateOfApplicabilityFilter", defaultValue: { snapshotId: null } }
  ) {
      statesOfApplicability(
          first: $first
          after: $after
          last: $last
          before: $before
          orderBy: $order
          filter: $filter
      ) @connection(key: "StatesOfApplicabilityPage_statesOfApplicability", filters: ["filter"]) {
          __id
          edges {
              node {
                  id
                  ...StateOfApplicabilityRowFragment
              }
          }
      }
  }
`;

export default function StatesOfApplicabilityPage({
  queryRef,
}: {
  queryRef: PreloadedQuery<StatesOfApplicabilityPageQuery>;
}) {
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("States of Applicability"));

  const { organization } = usePreloadedQuery(statesOfApplicabilityPageQuery, queryRef);

  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const {
    data: { statesOfApplicability },
    loadNext,
    hasNext,
    refetch,
    isLoadingNext,
  } = usePaginationFragment<
    StatesOfApplicabilityPagePaginationQuery,
    StatesOfApplicabilityPageFragment$key
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
        title={__("States of Applicability")}
        description={__(
          "Manage states of applicability for your organization's frameworks.",
        )}
      >
        {!isSnapshotMode
          && organization.canCreateStateOfApplicability && (
          <CreateStateOfApplicabilityDialog
            connectionId={statesOfApplicability.__id}
          >
            <Button icon={IconPlusLarge}>
              {__("Add state of applicability")}
            </Button>
          </CreateStateOfApplicabilityDialog>
        )}
      </PageHeader>

      {statesOfApplicability && statesOfApplicability.edges.length > 0
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
                  {statesOfApplicability.edges.map(edge => (
                    <StateOfApplicabilityRow
                      key={edge.node.id}
                      fKey={edge.node}
                      connectionId={statesOfApplicability.__id}
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
                  {__("No states of applicability yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__(
                    "Create your first state of applicability to get started.",
                  )}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
