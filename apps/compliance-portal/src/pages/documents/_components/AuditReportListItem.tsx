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

import { graphql, useFragment } from "react-relay";

import { useExportAndOpen } from "../_lib/useExportAndOpen";

import type { AuditReportListItem_audit$key } from "./__generated__/AuditReportListItem_audit.graphql";
import type { AuditReportListItemExportMutation } from "./__generated__/AuditReportListItemExportMutation.graphql";
import { DocumentEntry } from "./DocumentEntry";

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

// A single audit report entry: the framework name, the report file name, and an
// access action that opens the exported report when the viewer is authorized.
// Renders nothing when the audit has no report file.
export function AuditReportListItem({ auditKey }: AuditReportListItemProps) {
  const audit = useFragment(auditReportListItemFragment, auditKey);
  const [openReport, isExporting] = useExportAndOpen<AuditReportListItemExportMutation>(
    exportReportMutation,
    response => response.exportReportPDF.data,
  );

  const report = audit.reportFile;
  if (report == null) {
    return null;
  }

  return (
    <DocumentEntry
      title={audit.framework.name}
      meta={report.fileName}
      isAuthorized={report.isUserAuthorized}
      requested={report.access?.status === "REQUESTED"}
      onView={() => openReport({ input: { reportId: report.id } })}
      isViewing={isExporting}
    />
  );
}
