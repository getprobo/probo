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

export type RegistryStatus = "CLOSED" | "FALSE_POSITIVE" | "IN_PROGRESS" | "MITIGATED" | "OPEN" | "RISK_ACCEPTED";

export const registryStatuses = [
  "OPEN",
  "IN_PROGRESS",
  "CLOSED",
  "RISK_ACCEPTED",
  "MITIGATED",
  "FALSE_POSITIVE",
] as const;

export const getStatusVariant = (status: RegistryStatus) => {
  switch (status) {
    case "OPEN":
      return "danger" as const;
    case "IN_PROGRESS":
      return "warning" as const;
    case "CLOSED":
      return "success" as const;
    case "MITIGATED":
      return "success" as const;
    case "RISK_ACCEPTED":
      return "neutral" as const;
    case "FALSE_POSITIVE":
      return "neutral" as const;
    default:
      return "neutral" as const;
  }
};

export const getStatusLabel = (status: RegistryStatus) => {
  switch (status) {
    case "OPEN":
      return "Open";
    case "IN_PROGRESS":
      return "In Progress";
    case "CLOSED":
      return "Closed";
    case "RISK_ACCEPTED":
      return "Risk Accepted";
    case "MITIGATED":
      return "Mitigated";
    case "FALSE_POSITIVE":
      return "False Positive";
    default:
      return status;
  }
};

export function getStatusOptions(__: Translator) {
  return registryStatuses.map((status) => ({
    value: status,
    label: __({
      "OPEN": "Open",
      "IN_PROGRESS": "In Progress",
      "CLOSED": "Closed",
      "RISK_ACCEPTED": "Risk Accepted",
      "MITIGATED": "Mitigated",
      "FALSE_POSITIVE": "False Positive",
    }[status]),
  }));
}
