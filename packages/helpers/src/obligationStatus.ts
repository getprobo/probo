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

export type ObligationStatus = "NON_COMPLIANT" | "PARTIALLY_COMPLIANT" | "COMPLIANT";

export const obligationStatuses = [
  "NON_COMPLIANT",
  "PARTIALLY_COMPLIANT",
  "COMPLIANT",
] as const;

export const getObligationStatusVariant = (status: ObligationStatus) => {
  switch (status) {
    case "NON_COMPLIANT":
      return "danger" as const;
    case "PARTIALLY_COMPLIANT":
      return "warning" as const;
    case "COMPLIANT":
      return "success" as const;
    default:
      return "neutral" as const;
  }
};

export const getObligationStatusLabel = (status: ObligationStatus) => {
  switch (status) {
    case "NON_COMPLIANT":
      return "Non-compliant";
    case "PARTIALLY_COMPLIANT":
      return "Partially compliant";
    case "COMPLIANT":
      return "Compliant";
    default:
      return status;
  }
};

export function getObligationStatusOptions(__: Translator) {
  return obligationStatuses.map((status) => ({
    value: status,
    label: __({
      "NON_COMPLIANT": "Non-compliant",
      "PARTIALLY_COMPLIANT": "Partially compliant",
      "COMPLIANT": "Compliant",
    }[status]),
  }));
}
