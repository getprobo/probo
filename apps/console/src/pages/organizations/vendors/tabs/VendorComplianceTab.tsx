import { useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";
import type { VendorComplianceTabFragment$key } from "./__generated__/VendorComplianceTabFragment.graphql";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  ActionDropdown,
  DropdownItem,
  Dropzone,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useFragment, useMutation, useRefetchableFragment } from "react-relay";
import type { VendorComplianceTabFragment_report$key } from "./__generated__/VendorComplianceTabFragment_report.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf, fileSize, formatDate } from "@probo/helpers";
import { SortableTable, SortableTh } from "/components/SortableTable";

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
    }
  }
`;

const uploadReportMutation = graphql`
  mutation VendorComplianceTabUploadReportMutation(
    $input: UploadVendorComplianceReportInput!
    $connections: [ID!]!
  ) {
    uploadVendorComplianceReport(input: $input) {
      vendorComplianceReportEdge @appendEdge(connections: $connections) {
        node {
          id
          ...VendorComplianceTabFragment_report
        }
      }
    }
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
    vendor: VendorComplianceTabFragment$key & { name: string; id: string };
  }>();
  const [data, refetch] = useRefetchableFragment(
    complianceReportsFragment,
    vendor
  );
  const connectionId = data.complianceReports.__id;
  const reports = data.complianceReports.edges.map((edge) => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  usePageTitle(vendor.name + " - " + __("Compliance reports"));

  const [mutate, isMutating] = useMutation(uploadReportMutation);

  const handleDrop = (files: File[]) => {
    const reportDate = new Date().toISOString();
    for (const file of files) {
      mutate({
        variables: {
          connections: [connectionId],
          input: {
            vendorId: vendor.id,
            reportDate,
            reportName: file.name,
            file: null,
          },
        },
        uploadables: {
          "input.file": file,
        },
      });
    }
  };

  return (
    <div className="space-y-6">
      {!isSnapshotMode && (
        <Dropzone
          description={__("Only PDF files up to 10MB are allowed")}
          isUploading={isMutating}
          onDrop={handleDrop}
          accept={{
            "application/pdf": [".pdf"],
          }}
          maxSize={10}
        />
      )}
      <SortableTable refetch={refetch}>
        <Thead>
          <Tr>
            <Th>{__("Report name")}</Th>
            <SortableTh field="REPORT_DATE">{__("Report date")}</SortableTh>
            <Th>{__("Valid until")}</Th>
            <Th>{__("File size")}</Th>
            {!isSnapshotMode && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {reports.map((report) => (
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
    props.reportKey
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
            'This will permanently delete the report "%s". This action cannot be undone.'
          ),
          report.reportName
        ),
      }
    );
  };

  return (
    <Tr>
      <Td>{report.reportName}</Td>
      <Td>{formatDate(report.reportDate)}</Td>
      <Td>{formatDate(report.validUntil)}</Td>
      <Td>{fileSize(__, report.file?.size)}</Td>
      {!props.isSnapshotMode && (
        <Td width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              onClick={handleDelete}
              variant="danger"
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
