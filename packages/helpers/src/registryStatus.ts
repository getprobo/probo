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

export function getStatusOptions(t: Translator) {
  return registryStatuses.map((status) => ({
    value: status,
    label: t({
      "OPEN": "helpers.registryStatus.open",
      "IN_PROGRESS": "helpers.registryStatus.inProgress",
      "CLOSED": "helpers.registryStatus.closed",
      "RISK_ACCEPTED": "helpers.registryStatus.riskAccepted",
      "MITIGATED": "helpers.registryStatus.mitigated",
      "FALSE_POSITIVE": "helpers.registryStatus.falsePositive",
    }[status]),
  }));
}
