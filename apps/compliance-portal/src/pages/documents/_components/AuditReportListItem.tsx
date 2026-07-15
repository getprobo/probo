// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { useMutation } from "#/lib/relay/useMutation";

import { openExportedFile } from "../_lib/openExportedFile";

import type { AuditReportListItem_audit$key } from "./__generated__/AuditReportListItem_audit.graphql";
import type { AuditReportListItemExportMutation } from "./__generated__/AuditReportListItemExportMutation.graphql";
import { DocumentAccessAction } from "./DocumentAccessAction";
import { documentListItem } from "./variants";

const auditReportListItemFragment = graphql`
  fragment AuditReportListItem_audit on Audit @throwOnFieldError {
    framework {
      name
    }
    reportFile {
      id
      fileName
      isUserAuthorized
      access {
        status
      }
    }
  }
`;

const exportReportMutation = graphql`
  mutation AuditReportListItemExportMutation($input: ExportReportPDFInput!) {
    exportReportPDF(input: $input) {
      data
    }
  }
`;

interface AuditReportListItemProps {
  auditKey: AuditReportListItem_audit$key;
}

// A single audit report row: the framework name, the report file name, and an
// access action that opens the exported report when the viewer is authorized.
// Renders nothing when the audit has no report file.
export function AuditReportListItem({ auditKey }: AuditReportListItemProps) {
  const audit = useFragment(auditReportListItemFragment, auditKey);
  const [exportReport, isExporting] = useMutation<AuditReportListItemExportMutation>(exportReportMutation);
  const { root, content } = documentListItem();

  const report = audit.reportFile;
  if (report == null) {
    return null;
  }

  const handleView = () => {
    exportReport({
      variables: { input: { reportId: report.id } },
      onCompleted: response => openExportedFile(response.exportReportPDF.data),
    }).catch(() => {
      // The mutation failure is already surfaced through a toast.
    });
  };

  return (
    <div className={root()}>
      <div className={content()}>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {audit.framework.name}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {report.fileName}
        </Text>
      </div>
      <DocumentAccessAction
        isAuthorized={report.isUserAuthorized}
        requested={report.access?.status === "REQUESTED"}
        onView={handleView}
        isViewing={isExporting}
      />
    </div>
  );
}
