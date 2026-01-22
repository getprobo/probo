import {
  formatDate,
  getObligationStatusLabel,
  getObligationStatusVariant,
  promisifyMutation,
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
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type { ObligationGraphListQuery } from "#/__generated__/core/ObligationGraphListQuery.graphql";
import type {
  ObligationsPageFragment$data,
  ObligationsPageFragment$key,
} from "#/__generated__/core/ObligationsPageFragment.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  deleteObligationMutation,
  obligationsQuery,
} from "../../../hooks/graph/ObligationGraph";

import { CreateObligationDialog } from "./dialogs/CreateObligationDialog";

type Obligation
  = ObligationsPageFragment$data["obligations"]["edges"][number]["node"];

interface ObligationsPageProps {
  queryRef: PreloadedQuery<ObligationGraphListQuery>;
}

const obligationsPageFragment = graphql`
  fragment ObligationsPageFragment on Organization
  @refetchable(queryName: "ObligationsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    id
    obligations(
      first: $first
      after: $after
      filter: { snapshotId: $snapshotId }
    ) @connection(key: "ObligationsPage_obligations", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          id
          snapshotId
          sourceId
          area
          source
          requirement
          status
          lastReviewDate
          dueDate
          actionsToBeImplemented
          regulator
          owner {
            id
            fullName
          }
          createdAt
          updatedAt
          canUpdate: permission(action: "core:obligation:update")
          canDelete: permission(action: "core:obligation:delete")
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export default function ObligationsPage({ queryRef }: ObligationsPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(__("Obligations"));

  const organization = usePreloadedQuery(obligationsQuery, queryRef);

  const {
    data: obligationsData,
    loadNext,
    hasNext,
  } = usePaginationFragment(
    obligationsPageFragment,
    organization.node as ObligationsPageFragment$key,
  );

  const connectionId = obligationsData?.obligations?.__id || "";
  const obligations: Obligation[]
    = obligationsData?.obligations?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction
    = !isSnapshotMode
      && obligations.some(({ canUpdate, canDelete }) => canDelete || canUpdate);

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader
        title={__("Obligations")}
        description={__("Manage your organization's obligations.")}
      >
        {!snapshotId && organization.node.canCreateObligation && (
          <CreateObligationDialog
            organizationId={organizationId}
            connection={connectionId}
          >
            <Button icon={IconPlusLarge}>{__("Add obligation")}</Button>
          </CreateObligationDialog>
        )}
      </PageHeader>

      {obligations.length === 0
        ? (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No obligations yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__("Create your first obligation to get started.")}
                </p>
              </div>
            </Card>
          )
        : (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Area")}</Th>
                    <Th>{__("Source")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Owner")}</Th>
                    <Th>{__("Due Date")}</Th>
                    {hasAnyAction && <Th>{__("Actions")}</Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {obligations.map(obligation => (
                    <ObligationRow
                      key={obligation.id}
                      obligation={obligation}
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

function ObligationRow({
  obligation,
  connectionId,
  snapshotId,
  hasAnyAction,
}: {
  obligation: Obligation;
  connectionId: string;
  snapshotId?: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [deleteObligation] = useMutation(deleteObligationMutation);
  const confirm = useConfirm();
  const isSnapshotMode = Boolean(snapshotId);

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteObligation)({
          variables: {
            input: {
              obligationId: obligation.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: __(
          "This will permanently delete this obligation. This action cannot be undone.",
        ),
      },
    );
  };

  const detailsUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/obligations/${obligation.id}`
    : `/organizations/${organizationId}/obligations/${obligation.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>{obligation.area || "-"}</Td>
      <Td>{obligation.source || "-"}</Td>
      <Td>
        <Badge
          variant={getObligationStatusVariant(
            obligation.status || "NON_COMPLIANT",
          )}
        >
          {getObligationStatusLabel(obligation.status || "NON_COMPLIANT")}
        </Badge>
      </Td>
      <Td>{obligation.owner?.fullName || "-"}</Td>
      <Td>
        {obligation.dueDate
          ? (
              <time dateTime={obligation.dueDate}>
                {formatDate(obligation.dueDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{__("No due date")}</span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {obligation.canDelete && (
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
