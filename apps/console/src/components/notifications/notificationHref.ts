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

import { employeePortalHref } from "#/lib/employeePortalHref";

export type ConsoleNotificationType
  = | "DOCUMENT_SIGNATURE_REQUESTED"
    | "DOCUMENT_APPROVAL_REQUESTED"
    | "TASK_ASSIGNED"
    | "TASK_MENTIONED";

export type NotificationDestination
  = | { external: false; to: string }
    | { external: true; href: string };

export function notificationHref(
  organizationId: string,
  employeePortalId: string | null,
  type: ConsoleNotificationType,
  subjectId: string,
): NotificationDestination {
  if (type === "TASK_ASSIGNED" || type === "TASK_MENTIONED") {
    return {
      external: false,
      to: `/organizations/${organizationId}/governance/tasks/${subjectId}`,
    };
  }

  const section = type === "DOCUMENT_APPROVAL_REQUESTED"
    ? "approvals"
    : "signatures";

  if (employeePortalId != null) {
    return {
      external: true,
      href: employeePortalHref(employeePortalId, section, subjectId),
    };
  }

  return {
    external: false,
    to: `/organizations/${organizationId}/governance/documents/${subjectId}`,
  };
}
