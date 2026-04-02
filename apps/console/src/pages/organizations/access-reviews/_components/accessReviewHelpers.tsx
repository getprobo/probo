// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type BadgeVariant = "neutral" | "info" | "warning" | "success" | "danger";

export function statusBadgeVariant(status: string): BadgeVariant {
  switch (status) {
    case "DRAFT":
      return "neutral";
    case "IN_PROGRESS":
      return "info";
    case "PENDING_ACTIONS":
      return "warning";
    case "COMPLETED":
      return "success";
    case "FAILED":
    case "CANCELLED":
      return "danger";
    default:
      return "neutral";
  }
}

export function statusLabel(
  __: (key: string) => string,
  status: string,
): string {
  switch (status) {
    case "DRAFT":
      return __("Draft");
    case "IN_PROGRESS":
      return __("In progress");
    case "PENDING_ACTIONS":
      return __("Pending actions");
    case "COMPLETED":
      return __("Completed");
    case "FAILED":
      return __("Failed");
    case "CANCELLED":
      return __("Cancelled");
    default:
      return status;
  }
}

export function decisionBadgeVariant(decision: string): BadgeVariant {
  switch (decision) {
    case "APPROVED":
      return "success";
    case "REVOKE":
      return "danger";
    case "DEFER":
      return "warning";
    case "ESCALATE":
      return "info";
    default:
      return "neutral";
  }
}

export function decisionLabel(
  __: (key: string) => string,
  decision: string,
): string {
  switch (decision) {
    case "PENDING":
      return __("Pending");
    case "APPROVED":
      return __("Approved");
    case "REVOKE":
      return __("Revoked");
    case "DEFER":
      return __("Modified");
    case "ESCALATE":
      return __("Escalated");
    default:
      return decision;
  }
}

export function flagBadgeVariant(flag: string): BadgeVariant {
  switch (flag) {
    case "ORPHANED":
    case "TERMINATED_USER":
    case "CONTRACTOR_EXPIRED":
      return "danger";
    case "DORMANT":
    case "EXCESSIVE":
    case "SOD_CONFLICT":
    case "PRIVILEGED_ACCESS":
    case "ROLE_CREEP":
    case "ROLE_MISMATCH":
      return "warning";
    case "NO_BUSINESS_JUSTIFICATION":
    case "OUT_OF_DEPARTMENT":
    case "SHARED_ACCOUNT":
    case "INACTIVE":
    case "NEW":
      return "info";
    default:
      return "neutral";
  }
}

export const flagGroups = [
  {
    label: "Account",
    flags: [
      { value: "ORPHANED" as const, label: "Orphan account" },
      { value: "DORMANT" as const, label: "Dormant" },
      { value: "TERMINATED_USER" as const, label: "Terminated user" },
      { value: "CONTRACTOR_EXPIRED" as const, label: "Contractor expired" },
    ],
  },
  {
    label: "Privileges",
    flags: [
      { value: "EXCESSIVE" as const, label: "Excessive privileges" },
      { value: "SOD_CONFLICT" as const, label: "SoD conflict" },
      { value: "PRIVILEGED_ACCESS" as const, label: "Privileged access" },
      { value: "ROLE_CREEP" as const, label: "Role creep" },
    ],
  },
  {
    label: "Anomaly",
    flags: [
      { value: "NO_BUSINESS_JUSTIFICATION" as const, label: "No justification" },
      { value: "OUT_OF_DEPARTMENT" as const, label: "Out of department" },
      { value: "SHARED_ACCOUNT" as const, label: "Shared account" },
    ],
  },
];

export function flagLabel(flag: string): string {
  for (const group of flagGroups) {
    for (const f of group.flags) {
      if (f.value === flag) return f.label;
    }
  }
  if (flag === "NONE") return "None";
  // Legacy flag values not shown in the grouped dropdown
  if (flag === "INACTIVE") return "Inactive";
  if (flag === "ROLE_MISMATCH") return "Role mismatch";
  if (flag === "NEW") return "New";
  return flag;
}

export function formatStatus(status: string): string {
  return status.replace(/_/g, " ");
}

export function NotAvailable() {
  return (
    <span className="text-xs text-txt-tertiary">N/A</span>
  );
}
