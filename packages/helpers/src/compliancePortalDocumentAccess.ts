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

import type { CompliancePortalDocumentAccessStatus } from "@probo/coredata";

export function getCompliancePortalDocumentAccessStatusBadgeVariant(status: CompliancePortalDocumentAccessStatus) {
  switch (status) {
    case "REQUESTED":
      return "warning" as const;
    case "GRANTED":
      return "success" as const;
    case "REJECTED":
    case "REVOKED":
      return "danger" as const;
  }
}

export function getCompliancePortalDocumentAccessStatusLabel(status: CompliancePortalDocumentAccessStatus, t: (key: string) => string) {
  switch (status) {
    case "REQUESTED":
      return t("helpers.compliancePortalDocumentAccessStatus.requested");
    case "GRANTED":
      return t("helpers.compliancePortalDocumentAccessStatus.granted");
    case "REJECTED":
      return t("helpers.compliancePortalDocumentAccessStatus.rejected");
    case "REVOKED":
      return t("helpers.compliancePortalDocumentAccessStatus.revoked");
  }
}

interface ICompliancePortalDocumentAccessInfo {
  variant: string;
  type: string;
  persisted: boolean;
  name: string,
  typeLabel: string,
  category: string;
  id: string;
  status: CompliancePortalDocumentAccessStatus;
}

export type CompliancePortalDocumentAccessInfo = ICompliancePortalDocumentAccessInfo & (
  {
    variant: "info",
    type: "document",
  } | {
    variant: "success",
    type: "report",
  } | {
    variant: "highlight",
    type: "file",
  }
)
