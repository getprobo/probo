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

import { parseDate } from "@probo/helpers";
import {
  IconCircleCheck,
  IconCircleQuestionmark,
  IconCircleX,
} from "@probo/ui";

const MS_PER_SECOND = 1000;
const MS_PER_MINUTE = MS_PER_SECOND * 60;
const MS_PER_HOUR = MS_PER_MINUTE * 60;
const MS_PER_DAY = MS_PER_HOUR * 24;

// Compact elapsed age for dense list cells (e.g. "now", "5s", "20m", "4h", "1d", "1mo", "2y").
function shortAgeFormat(date: string, now: Date = new Date()): string {
  const elapsedMs = Math.max(0, now.getTime() - parseDate(date).getTime());
  const seconds = Math.floor(elapsedMs / MS_PER_SECOND);

  if (seconds < 60) {
    return seconds === 0 ? "now" : `${seconds}s`;
  }

  const minutes = Math.floor(elapsedMs / MS_PER_MINUTE);
  if (minutes < 60) {
    return `${minutes}m`;
  }

  const hours = Math.floor(elapsedMs / MS_PER_HOUR);
  if (hours < 24) {
    return `${hours}h`;
  }

  const days = Math.floor(elapsedMs / MS_PER_DAY);
  if (days < 30) {
    return `${days}d`;
  }
  if (days < 365) {
    return `${Math.floor(days / 30)}mo`;
  }
  return `${Math.floor(days / 365)}y`;
}

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
    case "CANCELLED":
      return "danger";
    default:
      return "neutral";
  }
}

export function isCampaignDeletableStatus(status: string): boolean {
  return status !== "IN_PROGRESS";
}

export function statusLabel(
  t: (key: string) => string,
  status: string,
): string {
  switch (status) {
    case "DRAFT":
      return t("accessReviewCampaignsPage.status.draft");
    case "IN_PROGRESS":
      return t("accessReviewCampaignsPage.status.in_progress");
    case "PENDING_ACTIONS":
      return t("accessReviewCampaignsPage.status.pending_actions");
    case "COMPLETED":
      return t("accessReviewCampaignsPage.status.completed");
    case "CANCELLED":
      return t("accessReviewCampaignsPage.status.cancelled");
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

export function NotAvailable() {
  return (
    <span className="text-xs text-txt-tertiary">N/A</span>
  );
}

export function MfaStatusIcon({
  status,
  label,
}: {
  status: string;
  label: string;
}) {
  if (status === "ENABLED") {
    return (
      <span role="img" aria-label={label} title={label} className="inline-flex text-txt-success">
        <IconCircleCheck size={16} />
      </span>
    );
  }

  if (status === "DISABLED") {
    return (
      <span role="img" aria-label={label} title={label} className="inline-flex text-txt-danger">
        <IconCircleX size={16} />
      </span>
    );
  }

  return (
    <span role="img" aria-label={label} title={label} className="inline-flex text-txt-tertiary">
      <IconCircleQuestionmark size={16} />
    </span>
  );
}

export function AuthMethodStatus({
  method,
  label,
  unknownLabel,
}: {
  method: string;
  label: string;
  unknownLabel: string;
}) {
  if (method === "UNKNOWN") {
    return (
      <span
        role="img"
        aria-label={unknownLabel}
        title={unknownLabel}
        className="inline-flex text-txt-tertiary"
      >
        <IconCircleQuestionmark size={16} />
      </span>
    );
  }

  return (
    <span
      aria-label={label}
      title={label}
      className="max-w-full truncate text-center text-xs text-txt-primary"
    >
      {label}
    </span>
  );
}

// Spelled out rather than a check/cross icon: next to MFA a green check reads as
// "compliant", but holding admin rights is the risk signal a reviewer looks for.
export function AdminStatus({
  isAdmin,
  trueLabel,
  falseLabel,
  unknownLabel,
}: {
  isAdmin: boolean | null | undefined;
  trueLabel: string;
  falseLabel: string;
  unknownLabel: string;
}) {
  if (isAdmin == null) {
    return (
      <span role="img" aria-label={unknownLabel} title={unknownLabel} className="inline-flex text-txt-tertiary">
        <IconCircleQuestionmark size={16} />
      </span>
    );
  }

  if (isAdmin) {
    return (
      <span aria-label={trueLabel} title={trueLabel} className="text-xs font-medium text-txt-warning">
        {trueLabel}
      </span>
    );
  }

  return (
    <span aria-label={falseLabel} title={falseLabel} className="text-xs text-txt-tertiary">
      {falseLabel}
    </span>
  );
}

export function LastLoginStatus({
  lastLogin,
  formatted,
  unknownLabel,
  compact = false,
}: {
  lastLogin: string | null | undefined;
  formatted: string;
  unknownLabel: string;
  compact?: boolean;
}) {
  if (lastLogin) {
    if (compact) {
      return (
        <span
          aria-label={formatted}
          title={formatted}
          className="text-xs tabular-nums text-txt-primary"
        >
          {shortAgeFormat(lastLogin)}
        </span>
      );
    }

    return (
      <span role="img" aria-label={formatted} className="inline-flex min-w-0 items-center gap-1.5" title={formatted}>
        <IconCircleCheck size={16} className="shrink-0 text-txt-tertiary" />
        <span className="truncate" aria-hidden="true">{formatted}</span>
      </span>
    );
  }

  return (
    <span role="img" aria-label={unknownLabel} title={unknownLabel} className="inline-flex text-txt-tertiary">
      <IconCircleQuestionmark size={16} />
    </span>
  );
}
