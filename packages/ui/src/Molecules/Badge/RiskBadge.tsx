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

import { useTranslation } from "react-i18next";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  level: number | string;
};

const badgeVariant = (level: string | number) => {
  if (typeof level === "number") {
    if (level >= 15) {
      level = "CRITICAL";
    } else if (level >= 8) {
      level = "HIGH";
    } else {
      level = "LOW";
    }
  }
  switch (level) {
    case "CRITICAL":
      return "danger";
    case "HIGH":
      return "warning";
    case "LOW":
      return "success";
    case "MEDIUM":
      return "info";
    default:
      return "neutral";
  }
};

export function RiskBadge({ level }: Props) {
  const { t } = useTranslation();
  const label = () => {
    if (typeof level === "number") {
      if (level >= 15) {
        return t("ui.risk.severity.high");
      }
      if (level >= 8) {
        return t("ui.risk.severity.medium");
      }
      return t("ui.risk.severity.low");
    }
    switch (level) {
      case "CRITICAL":
        return t("ui.risk.severity.critical");
      case "HIGH":
        return t("ui.risk.severity.high");
      case "LOW":
        return t("ui.risk.severity.low");
      case "MEDIUM":
        return t("ui.risk.severity.medium");
      case "NONE":
        return t("ui.risk.severity.none");
      default:
        return t("ui.risk.severity.low");
    }
  };
  return <Badge variant={badgeVariant(level)}>{label()}</Badge>;
}
