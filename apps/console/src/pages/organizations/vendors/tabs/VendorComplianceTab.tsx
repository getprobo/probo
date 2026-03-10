import { downloadFile, fileSize, formatDate, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconArrowDown,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { useFragment, useRefetchableFragment } from "react-relay";
import { useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { ComplianceReportListQuery } from "#/__generated__/core/ComplianceReportListQuery.graphql";
import type { VendorComplianceTabFragment$key } from "#/__generated__/core/VendorComplianceTabFragment.graphql";
import type { VendorComplianceTabFragment_report$key } from "#/__generated__/core/VendorComplianceTabFragment_report.graphql";
import type { VendorGraphNodeQuery$data } from "#/__generated__/core/VendorGraphNodeQuery.graphql";

import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { UploadComplianceReportDialog } from "../dialogs/UploadComplianceReportDialog";

import { UploadComplianceReportDialog } from "../dialogs/UploadComplianceReportDialog";

export const complianceReportsFragment = graphql`
  fragment VendorComplianceTabFragment on Vendor
  @refetchable(queryName: "ComplianceReportListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorComplianceReportOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    complianceReports(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorComplianceTabFragment_complianceReports") {
      __id
      edges {
        node {
          id
          canDelete: permission(action: "core:vendor-compliance-report:delete")
          ...VendorComplianceTabFragment_report
        }
      }
    }
  }
`;

const complianceReportFragment = graphql`
  fragment VendorComplianceTabFragment_report on VendorComplianceReport {
    id
    reportDate
    validUntil
    reportName
    file {
      fileName
      mimeType
      size
      downloadUrl
    }
    canDelete: permission(action: "core:vendor-compliance-report:delete")
  }
`;

const deleteReportMutation = graphql`
  mutation VendorComplianceTabDeleteReportMutation(
    $input: DeleteVendorComplianceReportInput!
    $connections: [ID!]!
  ) {
    deleteVendorComplianceReport(input: $input) {
      deletedVendorComplianceReportId @deleteEdge(connections: $connections)
    }
  }
`;

export default function VendorComplianceTab() {
  const { vendor } = useOutletContext<{
    vendor: VendorGraphNodeQuery$data["node"];
  }>();
  const [data, refetch] = useRefetchableFragment<
    ComplianceReportListQuery,
    VendorComplianceTabFragment$key
  >(complianceReportsFragment, vendor);
  const connectionId = data.complianceReports.__id;
  const reports = data.complianceReports.edges.map(edge => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  usePageTitle(vendor.name + " - " + __("Compliance reports"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Compliance reports")}
        description={__("Track vendor compliance certifications and reports.")}
      >
        {!isSnapshotMode && vendor.canUploadComplianceReport && (
          <UploadComplianceReportDialog
            vendorId={vendor.id}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>{__("Add report")}</Button>
          </UploadComplianceReportDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
        <Thead>
          <Tr>
            <Th>{__("Report name")}</Th>
            <SortableTh field="REPORT_DATE">{__("Report date")}</SortableTh>
            <Th>{__("Valid until")}</Th>
            <Th>{__("File size")}</Th>
            {!isSnapshotMode && reports.length > 0 && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {reports.map(report => (
            <ReportRow
              key={report.id}
              reportKey={report}
              connectionId={connectionId}
              isSnapshotMode={isSnapshotMode}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

type ReportRowProps = {
  reportKey: VendorComplianceTabFragment_report$key;
  connectionId: string;
  isSnapshotMode: boolean;
};

function ReportRow(props: ReportRowProps) {
  const { __ } = useTranslate();
  const report = useFragment<VendorComplianceTabFragment_report$key>(
    complianceReportFragment,
    props.reportKey,
  );
  const confirm = useConfirm();
  const [deleteReport] = useMutationWithToasts(deleteReportMutation, {
    successMessage: __("Report deleted successfully"),
    errorMessage: __("Failed to delete report"),
  });

  const handleDelete = () => {
    confirm(
      () =>
        deleteReport({
          variables: {
            connections: [props.connectionId],
            input: {
              reportId: report.id,
            },
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the report \"%s\". This action cannot be undone.",
          ),
          report.reportName,
        ),
      },
    );
  };

  return (
    <Tr>
      <Td>{report.reportName}</Td>
      <Td>{formatDate(report.reportDate)}</Td>
      <Td>{formatDate(report.validUntil)}</Td>
      <Td>{fileSize(__, report.file?.size ?? 0)}</Td>
      {!props.isSnapshotMode && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            {report.file?.downloadUrl && (
              <DropdownItem
                icon={IconArrowDown}
                onClick={() =>
                  downloadFile(
                    report.file!.downloadUrl,
                    report.file!.fileName,
                  )}
              >
                {__("Download")}
              </DropdownItem>
            )}
            {report.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                onClick={handleDelete}
                variant="danger"
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
