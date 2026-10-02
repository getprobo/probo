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

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { useTranslation } from "react-i18next";

import {
  type ConnectionPresentation,
  type ConnectionTone,
  connectionTone,
} from "#/pages/organizations/_lib/connectorStatus";

interface ConnectorConnectionBadgeProps {
  aggregated: boolean;
  aggregatedTotal: number | null;
  connectedCount: number;
  tone: ConnectionTone | "sand";
  solo: ConnectionPresentation | null;
}

export function ConnectorConnectionBadge({
  aggregated,
  aggregatedTotal,
  connectedCount,
  tone,
  solo,
}: ConnectorConnectionBadgeProps) {
  const { t } = useTranslation("organizations/settings/integrations");

  if (aggregated && aggregatedTotal != null) {
    return (
      <Badge
        variant="soft"
        color={tone === "sand" ? "neutral" : tone}
        size={1}
      >
        {t("listPage.connectedCount", {
          connected: connectedCount,
          total: aggregatedTotal,
        })}
      </Badge>
    );
  }

  if (!aggregated && solo != null) {
    return (
      <Badge
        variant="soft"
        color={connectionTone(solo.status)}
        size={1}
      >
        {t(`detailsPage.status.${solo.status}`)}
      </Badge>
    );
  }

  return null;
}
