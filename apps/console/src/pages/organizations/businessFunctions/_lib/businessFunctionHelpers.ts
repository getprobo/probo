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

export type BusinessFunctionClassification
  = "CRITICAL" | "IMPORTANT" | "SECONDARY" | "STANDARD";

type ClassificationBadgeVariant
  = "success" | "warning" | "danger" | "info" | "neutral" | "outline" | "highlight";

export function getClassificationLabel(
  classification: BusinessFunctionClassification,
  __: (s: string) => string,
): string {
  switch (classification) {
    case "CRITICAL":
      return __("Critical");
    case "IMPORTANT":
      return __("Important");
    case "SECONDARY":
      return __("Secondary");
    case "STANDARD":
      return __("Standard");
    default:
      return classification;
  }
}

export function getClassificationVariant(
  classification: BusinessFunctionClassification,
): ClassificationBadgeVariant {
  switch (classification) {
    case "CRITICAL":
      return "danger";
    case "IMPORTANT":
      return "warning";
    case "SECONDARY":
      return "neutral";
    case "STANDARD":
      return "success";
    default:
      return "neutral";
  }
}

export function businessFunctionClassificationOptions(__: (s: string) => string) {
  return ([
    "CRITICAL",
    "IMPORTANT",
    "SECONDARY",
    "STANDARD",
  ] as const).map(value => ({
    value,
    label: getClassificationLabel(value, __),
  }));
}

export function durationMinutesHelperText(__: (s: string) => string): string {
  return __("Common values: 1h = 60, 12h = 720, 24h = 1440, 3d = 4320 minutes");
}

export const BusinessFunctionsConnectionKey = "BusinessFunctionsPage_businessFunctions";

export const emptyBusinessFunctionFilter = {
  classification: null,
  ownerId: null,
  cifOnly: null,
};

export type BusinessFunctionListFilter = {
  classification: BusinessFunctionClassification | null;
  ownerId: string | null;
  cifOnly: boolean | null;
};

/** Relay connection filter keys a business function may appear under in the list. */
export function businessFunctionListConnectionFilters(businessFunction: {
  classification?: string | null;
  owner?: { id?: string | null } | null;
}): BusinessFunctionListFilter[] {
  const classification = (businessFunction.classification ?? null) as
    | BusinessFunctionClassification
    | null;
  const ownerId = businessFunction.owner?.id ?? null;
  const isCif = classification === "CRITICAL" || classification === "IMPORTANT";

  const filters: BusinessFunctionListFilter[] = [
    emptyBusinessFunctionFilter,
    { classification, ownerId: null, cifOnly: null },
  ];

  if (ownerId) {
    filters.push(
      { classification: null, ownerId, cifOnly: null },
      { classification, ownerId, cifOnly: null },
    );
  }

  if (isCif) {
    filters.push({ classification: null, ownerId: null, cifOnly: true });
    filters.push({ classification, ownerId: null, cifOnly: true });
    if (ownerId) {
      filters.push(
        { classification: null, ownerId, cifOnly: true },
        { classification, ownerId, cifOnly: true },
      );
    }
  }

  return filters;
}
