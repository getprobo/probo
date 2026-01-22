import {
  formatDate,
  getSnapshotTypeLabel,
  getSnapshotTypeUrlPath,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
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
import {
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { SnapshotGraphListQuery } from "#/__generated__/core/SnapshotGraphListQuery.graphql";
import type {
  SnapshotsPageFragment$data,
  SnapshotsPageFragment$key,
} from "#/__generated__/core/SnapshotsPageFragment.graphql";
import { snapshotsQuery, useDeleteSnapshot } from "#/hooks/graph/SnapshotGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import SnapshotFormDialog from "./dialog/SnapshotFormDialog";

type Props = {
  queryRef: PreloadedQuery<SnapshotGraphListQuery>;
};

const snapshotsFragment = graphql`
  fragment SnapshotsPageFragment on Organization {
    snapshots(first: 100)
      @connection(key: "SnapshotsGraphListQuery__snapshots") {
      __id
      edges {
        node {
          id
          name
          description
          type
          createdAt
          canDelete: permission(action: "core:snapshot:delete")
        }
      }
    }
  }
`;

export default function SnapshotsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const organization = usePreloadedQuery(
    snapshotsQuery,
    props.queryRef,
  ).organization;
  const data = useFragment<SnapshotsPageFragment$key>(
    snapshotsFragment,
    organization,
  );
  const connectionId = data.snapshots.__id;
  const snapshots = data.snapshots.edges.map(edge => edge.node);
  usePageTitle(__("Snapshots"));

  const hasAnyAction = snapshots.some(({ canDelete }) => canDelete);

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Snapshots")}
        description={__(
          "Snapshots capture point-in-time views of your organization's compliance state. Create snapshots to track progress over time.",
        )}
      >
        {organization.canCreateSnapshot && (
          <SnapshotFormDialog connection={connectionId}>
            <Button variant="primary" icon={IconPlusLarge}>
              {__("New snapshot")}
            </Button>
          </SnapshotFormDialog>
        )}
      </PageHeader>

      {snapshots.length > 0
        ? (
            <Table>
              <Thead>
                <Tr>
                  <Th>{__("Name")}</Th>
                  <Th>{__("Type")}</Th>
                  <Th>{__("Description")}</Th>
                  <Th>{__("Created")}</Th>
                  {hasAnyAction && <Th></Th>}
                </Tr>
              </Thead>
              <Tbody>
                {snapshots.map(snapshot => (
                  <SnapshotRow
                    key={snapshot.id}
                    snapshot={snapshot}
                    connectionId={connectionId}
                    organizationId={organizationId}
                    hasAnyAction={hasAnyAction}
                  />
                ))}
              </Tbody>
            </Table>
          )
        : (
            <div className="text-center py-12">
              <h3 className="text-lg font-medium text-txt-secondary mb-2">
                {__("No snapshots yet")}
              </h3>
              <p className="text-txt-tertiary mb-6">
                {__("Create your first snapshot to get started")}
              </p>
            </div>
          )}
    </div>
  );
}

type SnapshotRowProps = {
  snapshot: NodeOf<SnapshotsPageFragment$data["snapshots"]>;
  connectionId: string;
  organizationId: string;
  hasAnyAction: boolean;
};

function SnapshotRow(props: SnapshotRowProps) {
  const { __ } = useTranslate();
  const deleteSnapshot = useDeleteSnapshot(props.snapshot, props.connectionId);
  const typePath = getSnapshotTypeUrlPath(props.snapshot.type);

  return (
    <Tr
      to={`/organizations/${props.organizationId}/snapshots/${props.snapshot.id}${typePath}`}
    >
      <Td className="font-medium">{props.snapshot.name}</Td>
      <Td>
        <Badge variant="neutral">
          {getSnapshotTypeLabel(__, props.snapshot.type)}
        </Badge>
      </Td>
      <Td className="text-txt-secondary">
        {props.snapshot.description || __("No description")}
      </Td>
      <Td className="text-txt-tertiary">
        {formatDate(props.snapshot.createdAt)}
      </Td>
      {props.hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {props.snapshot.canDelete && (
              <DropdownItem
                onClick={deleteSnapshot}
                variant="danger"
                icon={IconTrashCan}
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
