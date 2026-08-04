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

export type MalaysiaPDPABreachStatus
  = | "OPEN"
    | "ASSESSING"
    | "CONTAINED"
    | "CLOSED";

export type BadgeVariant
  = | "success"
    | "warning"
    | "danger"
    | "info"
    | "neutral";

export const MALAYSIA_PDPA_BREACH_CONNECTION_KEY
  = "MalaysiaPDPABreachesPage__malaysiaPDPABreachIncidents";
export const MALAYSIA_PDPA_BREACH_HISTORY_CONNECTION_KEY
  = "MalaysiaPDPABreachStatusHistorySection__statusHistory";

export function getBreachStatusBadgeVariant(
  status: MalaysiaPDPABreachStatus,
): BadgeVariant {
  switch (status) {
    case "OPEN":
      return "danger";
    case "ASSESSING":
      return "warning";
    case "CONTAINED":
      return "info";
    case "CLOSED":
      return "neutral";
  }
}

export function getBreachDecisionBadgeVariant(decision: string): BadgeVariant {
  switch (decision) {
    case "NOT_REQUIRED":
      return "success";
    case "COMMISSIONER_ONLY":
      return "warning";
    case "COMMISSIONER_AND_DATA_SUBJECTS":
      return "danger";
    default:
      return "neutral";
  }
}

export function getAllowedBreachStatusTransitions(
  status: MalaysiaPDPABreachStatus,
): MalaysiaPDPABreachStatus[] {
  switch (status) {
    case "OPEN":
      return ["ASSESSING", "CONTAINED"];
    case "ASSESSING":
      return ["OPEN", "CONTAINED"];
    case "CONTAINED":
      return ["ASSESSING", "CLOSED"];
    case "CLOSED":
      return ["ASSESSING"];
  }
}
