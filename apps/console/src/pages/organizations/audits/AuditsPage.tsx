import {
  formatDate,
  getAuditStateLabel,
  getAuditStateVariant,
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
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditGraphListQuery } from "#/__generated__/core/AuditGraphListQuery.graphql";
import type {
  AuditsPageFragment$data,
  AuditsPageFragment$key,
} from "#/__generated__/core/AuditsPageFragment.graphql";
import { SortableTable } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { auditsQuery, useDeleteAudit } from "../../../hooks/graph/AuditGraph";

import { CreateAuditDialog } from "./dialogs/CreateAuditDialog";

const paginatedAuditsFragment = graphql`
  fragment AuditsPageFragment on Organization
  @refetchable(queryName: "AuditsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "AuditOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    audits(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "AuditsPage_audits") {
      __id
      edges {
        node {
          id
          name
          validFrom
          validUntil
          report {
            id
            filename
          }
          state
          framework {
            id
            name
          }
          createdAt
          canUpdate: permission(action: "core:audit:update")
          canDelete: permission(action: "core:audit:delete")
        }
      }
    }
  }
`;

type AuditEntry = NodeOf<AuditsPageFragment$data["audits"]>;

type Props = {
  queryRef: PreloadedQuery<AuditGraphListQuery>;
};

export default function AuditsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery(auditsQuery, props.queryRef);
  const pagination = usePaginationFragment(
    paginatedAuditsFragment,
    data.node as AuditsPageFragment$key,
  );
  const audits = pagination.data.audits?.edges?.map(edge => edge.node) ?? [];
  const connectionId = pagination.data.audits.__id;

  usePageTitle(__("Audits"));

  const hasAnyAction = audits.some(
    audit => audit.canDelete || audit.canUpdate,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Audits")}
        description={__(
          "Manage your organization's compliance audits and their progress.",
        )}
      >
        {data.node.canCreateAudit && (
          <CreateAuditDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add audit")}</Button>
          </CreateAuditDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination} pageSize={10}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Framework")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Valid From")}</Th>
            <Th>{__("Valid Until")}</Th>
            <Th>{__("Report")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {audits.map(entry => (
            <AuditRow
              key={entry.id}
              entry={entry}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function AuditRow({
  entry,
  connectionId,
  hasAnyAction,
}: {
  entry: AuditEntry;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deleteAudit = useDeleteAudit(entry, connectionId);

  return (
    <Tr to={`/organizations/${organizationId}/audits/${entry.id}`}>
      <Td>{entry.name || __("Untitled")}</Td>
      <Td>{entry.framework?.name ?? __("Unknown Framework")}</Td>
      <Td>
        <Badge variant={getAuditStateVariant(entry.state)}>
          {getAuditStateLabel(__, entry.state)}
        </Badge>
      </Td>
      <Td>{formatDate(entry.validFrom) || __("Not set")}</Td>
      <Td>{formatDate(entry.validUntil) || __("Not set")}</Td>
      <Td>
        {entry.report
          ? (
              <div className="flex flex-col">
                <Badge variant="success">{__("Uploaded")}</Badge>
              </div>
            )
          : (
              <Badge variant="neutral">{__("Not uploaded")}</Badge>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {entry.canDelete && (
              <DropdownItem
                onClick={deleteAudit}
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
