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

import type {
  NonconformitiesPageFragment$data,
  NonconformitiesPageFragment$key,
} from "#/__generated__/core/NonconformitiesPageFragment.graphql";
import type { NonconformityGraphListQuery } from "#/__generated__/core/NonconformityGraphListQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  deleteNonconformityMutation,
  NonconformitiesConnectionKey,
  nonconformitiesQuery,
} from "../../../hooks/graph/NonconformityGraph";

import { CreateNonconformityDialog } from "./dialogs/CreateNonconformityDialog";

type Nonconformity
  = NonconformitiesPageFragment$data["nonconformities"]["edges"][number]["node"];

interface NonconformitiesPageProps {
  queryRef: PreloadedQuery<NonconformityGraphListQuery>;
}

const nonconformitiesPageFragment = graphql`
  fragment NonconformitiesPageFragment on Organization
  @refetchable(queryName: "NonconformitiesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    nonconformities(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    )
      @connection(
        key: "NonconformitiesPage_nonconformities"
        filters: ["filter"]
      ) {
      __id
      totalCount
      edges {
        node {
          id
          referenceId
          snapshotId
          description
          status
          dateIdentified
          dueDate
          rootCause
          correctiveAction
          effectivenessCheck
          audit {
            id
            name
            framework {
              id
              name
            }
          }
          owner {
            id
            fullName
          }
          createdAt
          updatedAt
          canUpdate: permission(action: "core:nonconformity:update")
          canDelete: permission(action: "core:nonconformity:delete")
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export default function NonconformitiesPage({
  queryRef,
}: NonconformitiesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Nonconformities"));

  const organization = usePreloadedQuery(nonconformitiesQuery, queryRef);

  const {
    data: nonconformitiesData,
    loadNext,
    hasNext,
  } = usePaginationFragment(
    nonconformitiesPageFragment,
    organization.node as NonconformitiesPageFragment$key,
  );

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    NonconformitiesConnectionKey,
    { filter: { snapshotId: snapshotId || null } },
  );
  const nonconformities: Nonconformity[]
    = nonconformitiesData?.nonconformities?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction
    = !isSnapshotMode
      && nonconformities.some(({ canDelete, canUpdate }) => canDelete || canUpdate);

  return (
    <div className="space-y-6">
      {isSnapshotMode && <SnapshotBanner snapshotId={snapshotId!} />}
      <PageHeader
        title={__("Nonconformities")}
        description={__("Manage your organization's non conformities.")}
      >
        {!isSnapshotMode && organization.node.canCreateNonconformity && (
          <CreateNonconformityDialog
            organizationId={organizationId}
            connection={connectionId}
          >
            <Button icon={IconPlusLarge}>{__("Add nonconformity")}</Button>
          </CreateNonconformityDialog>
        )}
      </PageHeader>

      {nonconformities.length === 0
        ? (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No nonconformities yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__("Create your first nonconformity to get started.")}
                </p>
              </div>
            </Card>
          )
        : (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Reference ID")}</Th>
                    <Th>{__("Description")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Audit")}</Th>
                    <Th>{__("Owner")}</Th>
                    <Th>{__("Due Date")}</Th>
                    {hasAnyAction && <Th>{__("Actions")}</Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {nonconformities.map(nonconformity => (
                    <NonconformityRow
                      key={nonconformity.id}
                      nonconformity={nonconformity}
                      connectionId={connectionId}
                      isSnapshotMode={isSnapshotMode}
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
                    disabled={!hasNext}
                  >
                    {__("Load more")}
                  </Button>
                </div>
              )}
            </Card>
          )}
    </div>
  );
}

function NonconformityRow({
  nonconformity,
  connectionId,
  isSnapshotMode,
  snapshotId,
  hasAnyAction,
}: {
  nonconformity: Nonconformity;
  connectionId: string;
  isSnapshotMode: boolean;
  snapshotId?: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [deleteNonconformity] = useMutation(deleteNonconformityMutation);

  const nonconformityDetailUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/nonconformities/${nonconformity.id}`
    : `/organizations/${organizationId}/nonconformities/${nonconformity.id}`;

  const handleDeleteNonconformity = (nonconformity: Nonconformity) => {
    if (!connectionId) return;

    confirm(
      () => {
        return promisifyMutation(deleteNonconformity)({
          variables: {
            input: {
              nonconformityId: nonconformity.id,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __(
            "This will permanently delete the nonconformity %s. This action cannot be undone.",
          ),
          nonconformity.referenceId,
        ),
      },
    );
  };

  return (
    <Tr to={nonconformityDetailUrl}>
      <Td>
        <span className="font-mono text-sm">{nonconformity.referenceId}</span>
      </Td>
      <Td>
        <div className="min-w-0">
          <p className="whitespace-pre-wrap break-words">
            {nonconformity.description || __("No description")}
          </p>
        </div>
      </Td>
      <Td>
        <Badge variant={getStatusVariant(nonconformity.status)}>
          {getStatusLabel(nonconformity.status)}
        </Badge>
      </Td>
      <Td>
        {nonconformity.audit
          ? (
              nonconformity.audit.name
                ? (
                    `${nonconformity.audit.framework?.name} - ${nonconformity.audit.name}`
                  )
                : (
                    nonconformity.audit.framework?.name
                  )
            )
          : (
              <span className="text-txt-tertiary">{__("No audit")}</span>
            )}
      </Td>
      <Td>{nonconformity.owner.fullName}</Td>
      <Td>
        {nonconformity.dueDate
          ? (
              <time dateTime={nonconformity.dueDate}>
                {formatDate(nonconformity.dueDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{__("No due date")}</span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {nonconformity.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={() => handleDeleteNonconformity(nonconformity)}
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
