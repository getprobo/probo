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

/** `prefix` selects which namespace's `classifications.*` keys to use (each page owns its own copy). */
export function getClassificationLabel(
  classification: BusinessFunctionClassification,
  t: (key: string) => string,
  prefix: string,
): string {
  switch (classification) {
    case "CRITICAL":
      return t(`${prefix}.classifications.critical`);
    case "IMPORTANT":
      return t(`${prefix}.classifications.important`);
    case "SECONDARY":
      return t(`${prefix}.classifications.secondary`);
    case "STANDARD":
      return t(`${prefix}.classifications.standard`);
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

export function businessFunctionClassificationOptions(
  t: (key: string) => string,
  prefix: string,
) {
  return ([
    "CRITICAL",
    "IMPORTANT",
    "SECONDARY",
    "STANDARD",
  ] as const).map(value => ({
    value,
    label: getClassificationLabel(value, t, prefix),
  }));
}

export function durationMinutesHelperText(
  t: (key: string) => string,
  prefix: string,
): string {
  return t(`${prefix}.fields.durationHelper`);
}

type DurationTranslator = (key: string, options?: { count?: number }) => string;

export function formatContinuityMinutes(
  minutes: number,
  t: DurationTranslator,
): string {
  const totalMinutes = Number.isFinite(minutes) ? Math.max(0, Math.trunc(minutes)) : 0;
  const hours = Math.floor(totalMinutes / 60);
  const remainingMinutes = totalMinutes % 60;

  if (hours > 0 && remainingMinutes > 0) {
    return `${t("duration.hour", { count: hours })}, ${t("duration.min", { count: remainingMinutes })}`;
  }
  if (hours > 0) {
    return t("duration.hour", { count: hours });
  }
  return t("duration.min", { count: remainingMinutes });
}

export const BusinessFunctionsConnectionKey = "BusinessFunctionsPage_businessFunctions";

export const emptyBusinessFunctionFilter = {
  classification: null,
};

export type BusinessFunctionListFilter = {
  classification: BusinessFunctionClassification | null;
};

/** Relay connection filter keys a business function may appear under in the list. */
export function businessFunctionListConnectionFilters(businessFunction: {
  classification?: string | null;
}): BusinessFunctionListFilter[] {
  const classification = (businessFunction.classification ?? null) as
    | BusinessFunctionClassification
    | null;

  return [
    emptyBusinessFunctionFilter,
    { classification },
  ];
}
