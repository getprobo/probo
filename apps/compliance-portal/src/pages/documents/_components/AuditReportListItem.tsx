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

import { useLocalizedPath } from "#/lib/i18n/useLocale";

import { useRequestReportAccess } from "../_lib/useAccessRequest";

import type { AuditReportListItem_audit$key } from "./__generated__/AuditReportListItem_audit.graphql";
import { DocumentEntry } from "./DocumentEntry";

const auditReportListItemFragment = graphql`
  fragment AuditReportListItem_audit on Audit @throwOnFieldError {
    framework {
      name
    }
    reportFile {
      id
      alias
      fileName
      isUserAuthorized
      access {
        status
      }
    }
  }
`;

interface AuditReportListItemProps {
  auditKey: AuditReportListItem_audit$key;
}

// A single audit report entry: the framework name, the report file name, and an
// access action linking to the viewer when authorized. Renders nothing when the
// audit has no report file.
export function AuditReportListItem({ auditKey }: AuditReportListItemProps) {
  const localizedPath = useLocalizedPath();
  const audit = useFragment(auditReportListItemFragment, auditKey);
  const report = audit.reportFile;
  // Hook must run unconditionally; the empty id is never used when there is no
  // report file (the component returns null below).
  const { requestAccess, isRequesting } = useRequestReportAccess(report?.id ?? "");

  if (report == null) {
    return null;
  }

  return (
    <DocumentEntry
      title={audit.framework.name}
      meta={report.fileName}
      isAuthorized={report.isUserAuthorized}
      requested={report.access?.status === "REQUESTED"}
      viewHref={localizedPath(`/documents/${encodeURIComponent(report.alias ?? report.id)}`)}
      onGetAccess={requestAccess}
      isRequesting={isRequesting}
    />
  );
}
