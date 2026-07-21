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

export function getObligationStatusOptions(t: Translator) {
  return obligationStatuses.map((status) => ({
    value: status,
    label: t({
      "NON_COMPLIANT": "helpers.obligationStatus.nonCompliant",
      "PARTIALLY_COMPLIANT": "helpers.obligationStatus.partiallyCompliant",
      "COMPLIANT": "helpers.obligationStatus.compliant",
    }[status]),
  }));
}
