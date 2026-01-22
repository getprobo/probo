import {
  formatDate,
  getStatusLabel,
  getStatusVariant,
  promisifyMutation,
  sprintf,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
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
  useConfirm,
} from "@probo/ui";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type { ContinualImprovementGraphListQuery } from "#/__generated__/core/ContinualImprovementGraphListQuery.graphql";
import type {
  ContinualImprovementsPageFragment$data,
  ContinualImprovementsPageFragment$key,
} from "#/__generated__/core/ContinualImprovementsPageFragment.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import {
  ContinualImprovementsConnectionKey,
  continualImprovementsQuery,
  deleteContinualImprovementMutation,
} from "../../../hooks/graph/ContinualImprovementGraph";

import { CreateContinualImprovementDialog } from "./dialogs/CreateContinualImprovementDialog";

interface ContinualImprovementsPageProps {
  queryRef: PreloadedQuery<ContinualImprovementGraphListQuery>;
}

const continualImprovementsPageFragment = graphql`
  fragment ContinualImprovementsPageFragment on Organization
  @refetchable(queryName: "ContinualImprovementsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    continualImprovements(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    )
      @connection(
        key: "ContinualImprovementsPage_continualImprovements"
        filters: ["filter"]
      ) {
      __id
      totalCount
      edges {
        node {
          id
          snapshotId
          sourceId
          referenceId
          description
          source
          targetDate
          status
          priority
          owner {
            id
            fullName
          }
          createdAt
          updatedAt
          canUpdate: permission(action: "core:continual-improvement:update")
          canDelete: permission(action: "core:continual-improvement:delete")
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export default function ContinualImprovementsPage({
  queryRef,
}: ContinualImprovementsPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Continual Improvements"));

  const organization = usePreloadedQuery(continualImprovementsQuery, queryRef);

  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    ContinualImprovementGraphListQuery,
    ContinualImprovementsPageFragment$key
  >(continualImprovementsPageFragment, organization.node);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ContinualImprovementsConnectionKey,
    { filter: { snapshotId: snapshotId || null } },
  );
  const improvements
    = data?.continualImprovements?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction
    = !isSnapshotMode
      && improvements.some(({ canUpdate, canDelete }) => canUpdate || canDelete);

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader
        title={__("Continual Improvements")}
        description={__("Manage your continual improvements.")}
      >
        {!isSnapshotMode && organization.node.canCreateContinualImprovement && (
          <CreateContinualImprovementDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {__("Add continual improvement")}
            </Button>
          </CreateContinualImprovementDialog>
        )}
      </PageHeader>

      {improvements.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Reference ID")}</Th>
                    <Th>{__("Description")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Priority")}</Th>
                    <Th>{__("Owner")}</Th>
                    <Th>{__("Target Date")}</Th>
                    {hasAnyAction && <Th>{__("Actions")}</Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {improvements.map(improvement => (
                    <ImprovementRow
                      key={improvement.id}
                      improvement={improvement}
                      connectionId={connectionId}
                      snapshotId={snapshotId}
                      hasAnyAction={hasAnyAction}
                    />
                  ))}
                </Tbody>
              </Table>

              {hasNext && (
                <div className="p-4 border-t">
                  <Button
                    variant="secondary"
                    onClick={() => loadNext(10)}
                    disabled={isLoadingNext}
                  >
                    {isLoadingNext ? __("Loading...") : __("Load more")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No continual improvements yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__("Create your first continual improvement to get started.")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}

function ImprovementRow({
  improvement,
  connectionId,
  snapshotId,
  hasAnyAction,
}: {
  improvement: NodeOf<
    NonNullable<ContinualImprovementsPageFragment$data["continualImprovements"]>
  >;
  connectionId: string;
  snapshotId?: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [deleteImprovement] = useMutation(deleteContinualImprovementMutation);
  const confirm = useConfirm();
  const isSnapshotMode = Boolean(snapshotId);

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteImprovement)({
          variables: {
            input: {
              continualImprovementId: improvement.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the continual improvement entry %s. This action cannot be undone.",
          ),
          improvement.referenceId,
        ),
      },
    );
  };

  const detailsUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/continual-improvements/${improvement.id}`
    : `/organizations/${organizationId}/continual-improvements/${improvement.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <span className="font-mono text-sm">{improvement.referenceId}</span>
      </Td>
      <Td>{improvement.description || "-"}</Td>
      <Td>
        <Badge variant={getStatusVariant(improvement.status)}>
          {getStatusLabel(improvement.status)}
        </Badge>
      </Td>
      <Td>
        <Badge
          variant={
            improvement.priority === "HIGH"
              ? "danger"
              : improvement.priority === "MEDIUM"
                ? "warning"
                : "success"
          }
        >
          {improvement.priority === "HIGH"
            ? __("High")
            : improvement.priority === "MEDIUM"
              ? __("Medium")
              : __("Low")}
        </Badge>
      </Td>
      <Td>{improvement.owner?.fullName || "-"}</Td>
      <Td>
        {improvement.targetDate
          ? (
              <time dateTime={improvement.targetDate}>
                {formatDate(improvement.targetDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{__("No target date")}</span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {improvement.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
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
