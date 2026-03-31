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

import { useTranslate } from "@probo/i18n";

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
  const { __ } = useTranslate();
  const label = () => {
    if (typeof level === "number") {
      if (level >= 15) {
        return __("High");
      }
      if (level >= 8) {
        return __("Medium");
      }
      return __("Low");
    }
    switch (level) {
      case "CRITICAL":
        return __("Critical");
      case "HIGH":
        return __("High");
      case "LOW":
        return __("Low");
      case "MEDIUM":
        return __("Medium");
      case "NONE":
        return __("None");
      default:
        return __("Low");
    }
  };
  return <Badge variant={badgeVariant(level)}>{label()}</Badge>;
}
