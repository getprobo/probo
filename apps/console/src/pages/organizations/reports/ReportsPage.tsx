import {
  formatDate,
  getReportStateLabel,
  getReportStateVariant,
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

import type {
  ReportsPageFragment$data,
  ReportsPageFragment$key,
} from "#/__generated__/core/ReportsPageFragment.graphql";
import type { ReportsPageQuery } from "#/__generated__/core/ReportsPageQuery.graphql";
import { SortableTable } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { useDeleteReport } from "../../../hooks/graph/ReportGraph";

export const reportsPageQuery = graphql`
  query ReportsPageQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateReport: permission(action: "core:report:create")
        ...ReportsPageFragment
      }
    }
  }
`;

import { CreateReportDialog } from "./dialogs/CreateReportDialog";

const paginatedReportsFragment = graphql`
  fragment ReportsPageFragment on Organization
  @refetchable(queryName: "ReportsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "ReportOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    reports(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "ReportsPage_reports") {
      __id
      edges {
        node {
          id
          name
          frameworkType
          validFrom
          validUntil
          file {
            id
            fileName
          }
          state
          framework {
            id
            name
          }
          createdAt
          canUpdate: permission(action: "core:report:update")
          canDelete: permission(action: "core:report:delete")
        }
      }
    }
  }
`;

type ReportEntry = NodeOf<ReportsPageFragment$data["reports"]>;

type Props = {
  queryRef: PreloadedQuery<ReportsPageQuery>;
};

export default function ReportsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery(reportsPageQuery, props.queryRef);
  const pagination = usePaginationFragment(
    paginatedReportsFragment,
    data.node as ReportsPageFragment$key,
  );
  const reports = pagination.data.reports?.edges?.map(edge => edge.node) ?? [];
  const connectionId = pagination.data.reports.__id;

  usePageTitle(__("Reports"));

  const hasAnyAction = reports.some(
    report => report.canDelete || report.canUpdate,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Reports")}
        description={__(
          "Manage your organization's compliance reports and their progress.",
        )}
      >
        {data.node.canCreateReport && (
          <CreateReportDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add report")}</Button>
          </CreateReportDialog>
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
            <Th>{__("File")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {reports.map(entry => (
            <ReportRow
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

function ReportRow({
  entry,
  connectionId,
  hasAnyAction,
}: {
  entry: ReportEntry;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deleteReport = useDeleteReport(entry, connectionId);

  return (
    <Tr to={`/organizations/${organizationId}/reports/${entry.id}`}>
      <Td>{entry.name || __("Untitled")}</Td>
      <Td>
        {entry.framework?.name ?? __("Unknown Framework")}
        {entry.frameworkType && (
          <span className="text-txt-secondary ml-1">{entry.frameworkType}</span>
        )}
      </Td>
      <Td>
        <Badge variant={getReportStateVariant(entry.state)}>
          {getReportStateLabel(__, entry.state)}
        </Badge>
      </Td>
      <Td>{formatDate(entry.validFrom) || __("Not set")}</Td>
      <Td>{formatDate(entry.validUntil) || __("Not set")}</Td>
      <Td>
        {entry.file
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
                onClick={deleteReport}
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
