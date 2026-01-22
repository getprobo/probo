import { formatDate } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  Card,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useEffect } from "react";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type {
  StateOfApplicabilityGraphPaginatedFragment$data,
  StateOfApplicabilityGraphPaginatedFragment$key,
} from "/__generated__/core/StateOfApplicabilityGraphPaginatedFragment.graphql";
import type { StateOfApplicabilityGraphPaginatedQuery } from "/__generated__/core/StateOfApplicabilityGraphPaginatedQuery.graphql";
import type { StateOfApplicabilityListQuery } from "/__generated__/core/StateOfApplicabilityListQuery.graphql";
import { SnapshotBanner } from "/components/SnapshotBanner";
import {
  paginatedStateOfApplicabilityFragment,
  paginatedStateOfApplicabilityQuery,
  useDeleteStateOfApplicability,
} from "/hooks/graph/StateOfApplicabilityGraph";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { NodeOf } from "/types";

import { CreateStateOfApplicabilityDialog } from "./dialogs/CreateStateOfApplicabilityDialog";

type StateOfApplicability = NodeOf<
  StateOfApplicabilityGraphPaginatedFragment$data["statesOfApplicability"]
>;

export default function StatesOfApplicabilityPage({
  queryRef,
}: {
  queryRef: PreloadedQuery<StateOfApplicabilityGraphPaginatedQuery>;
}) {
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("States of Applicability"));

  const { organization }
    = usePreloadedQuery<StateOfApplicabilityGraphPaginatedQuery>(
      paginatedStateOfApplicabilityQuery,
      queryRef,
    );

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
    StateOfApplicabilityListQuery,
    StateOfApplicabilityGraphPaginatedFragment$key
  >(paginatedStateOfApplicabilityFragment, organization);

  // Refetch with snapshot filter when in snapshot mode
  useEffect(() => {
    if (snapshotId) {
      refetch(
        { filter: { snapshotId } },
        { fetchPolicy: "store-or-network" },
      );
    }
  }, [snapshotId, refetch]);

  const hasAnyAction
    = !isSnapshotMode
      && statesOfApplicability.edges
        .map(edge => edge.node)
        .some(({ canDelete }) => canDelete);

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
                    {hasAnyAction && <Th>{__("Actions")}</Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {statesOfApplicability.edges.map(edge => (
                    <StateOfApplicabilityRow
                      key={edge.node.id}
                      stateOfApplicability={edge.node}
                      connectionId={statesOfApplicability.__id}
                      hasAnyAction={hasAnyAction}
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

function StateOfApplicabilityRow({
  stateOfApplicability,
  connectionId,
  hasAnyAction,
}: {
  stateOfApplicability: StateOfApplicability;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const deleteStateOfApplicability = useDeleteStateOfApplicability(
    stateOfApplicability,
    connectionId,
  );

  const detailUrl = snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/states-of-applicability/${stateOfApplicability.id}`
    : `/organizations/${organizationId}/states-of-applicability/${stateOfApplicability.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{stateOfApplicability.name}</Td>
      <Td>
        <time dateTime={stateOfApplicability.createdAt}>
          {formatDate(stateOfApplicability.createdAt)}
        </time>
      </Td>
      <Td>{stateOfApplicability.controlsInfo?.totalCount ?? 0}</Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {stateOfApplicability.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  deleteStateOfApplicability();
                }}
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
