// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { downloadFile, fileSize, formatDate, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconArrowDown,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ThirdPartyComplianceReportRowDeleteMutation } from "#/__generated__/core/ThirdPartyComplianceReportRowDeleteMutation.graphql";
import type { ThirdPartyComplianceReportRow_report$key } from "#/__generated__/core/ThirdPartyComplianceReportRow_report.graphql";

const complianceReportRowFragment = graphql`
  fragment ThirdPartyComplianceReportRow_report on ThirdPartyComplianceReport {
    id
    reportDate
    validUntil
    reportName
    file {
      fileName
      size
      downloadUrl
    }
    canDelete: permission(action: "core:thirdParty-compliance-report:delete")
  }
`;

const deleteReportMutation = graphql`
  mutation ThirdPartyComplianceReportRowDeleteMutation(
    $input: DeleteThirdPartyComplianceReportInput!
    $connections: [ID!]!
  ) {
    deleteThirdPartyComplianceReport(input: $input) {
      deletedThirdPartyComplianceReportId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyComplianceReportRowProps {
  reportKey: ThirdPartyComplianceReportRow_report$key;
  connectionId: string;
}

export function ThirdPartyComplianceReportRow(
  props: ThirdPartyComplianceReportRowProps,
) {
  const { __ } = useTranslate();
  const report = useFragment(complianceReportRowFragment, props.reportKey);
  const confirm = useConfirm();
  const { toast } = useToast();
  const [deleteReport] = useMutation<ThirdPartyComplianceReportRowDeleteMutation>(
    deleteReportMutation,
  );

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteReport({
            variables: {
              connections: [props.connectionId],
              input: { reportId: report.id },
            },
            onCompleted() {
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete report"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
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
    </Tr>
  );
}
