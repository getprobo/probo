// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type Translator = (s: string) => string;

export const auditStates = [
  "NOT_STARTED",
  "IN_PROGRESS",
  "COMPLETED",
  "REJECTED",
  "OUTDATED",
] as const;

export function getAuditStateLabel(__: Translator, state: (typeof auditStates)[number]) {
  switch (state) {
    case "NOT_STARTED":
      return __("Not Started");
    case "IN_PROGRESS":
      return __("In Progress");
    case "COMPLETED":
      return __("Completed");
    case "REJECTED":
      return __("Rejected");
    case "OUTDATED":
      return __("Outdated");
    default:
      return __("Unknown");
  }
}

export function getAuditStateVariant(state: (typeof auditStates)[number]) {
  switch (state) {
    case "NOT_STARTED":
      return "neutral";
    case "IN_PROGRESS":
      return "info";
    case "COMPLETED":
      return "success";
    case "REJECTED":
      return "danger";
    case "OUTDATED":
      return "warning";
    default:
      return "neutral";
  }
}
