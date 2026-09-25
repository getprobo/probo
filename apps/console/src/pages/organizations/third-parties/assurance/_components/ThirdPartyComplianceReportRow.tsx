// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { downloadFile, formatError } from "@probo/helpers";
import { dateFormat, fileSize } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ThirdPartyComplianceReportRow_report$key } from "#/__generated__/core/ThirdPartyComplianceReportRow_report.graphql";
import type { ThirdPartyComplianceReportRowDeleteMutation } from "#/__generated__/core/ThirdPartyComplianceReportRowDeleteMutation.graphql";

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
  const { t, i18n } = useTranslation();
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
            onCompleted(_response, errors) {
              if (errors) {
                toast({
                  title: t("thirdPartyComplianceReportRow.messages.error"),
                  description: formatError(
                    t("thirdPartyComplianceReportRow.errors.delete"),
                    errors,
                  ),
                  variant: "error",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("thirdPartyComplianceReportRow.messages.error"),
                description: formatError(
                  t("thirdPartyComplianceReportRow.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("thirdPartyComplianceReportRow.deleteConfirmation", { name: report.reportName }),
      },
    );
  };

  return (
    <Tr>
      <Td>{report.reportName}</Td>
      <Td>{dateFormat(i18n.language, report.reportDate)}</Td>
      <Td>{dateFormat(i18n.language, report.validUntil)}</Td>
      <Td>{fileSize(report.file?.size ?? 0, t)}</Td>
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
              {t("thirdPartyComplianceReportRow.actions.download")}
            </DropdownItem>
          )}
          {report.canDelete && (
            <DropdownItem
              icon={IconTrashCan}
              onClick={handleDelete}
              variant="danger"
            >
              {t("thirdPartyComplianceReportRow.actions.delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </Td>
    </Tr>
  );
}
