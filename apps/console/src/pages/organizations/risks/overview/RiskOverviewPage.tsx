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

import { Card } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { RiskOverviewPageQuery } from "#/__generated__/core/RiskOverviewPageQuery.graphql";

export const riskOverviewPageQuery = graphql`
  query RiskOverviewPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        note
      }
    }
  }
`;

interface RiskOverviewPageProps {
  queryRef: PreloadedQuery<RiskOverviewPageQuery>;
}

export default function RiskOverviewPage({ queryRef }: RiskOverviewPageProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<RiskOverviewPageQuery>(riskOverviewPageQuery, queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const note = data.node.note?.trim();

  return (
    <div className="space-y-6">
      <Card className="space-y-4" padded>
        {note
          ? (
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {t("riskOverviewPage.fields.note")}
                </div>
                <div className="text-sm text-txt-secondary whitespace-pre-wrap">{note}</div>
              </div>
            )
          : (
              <p className="text-sm text-txt-secondary">
                {t("riskOverviewPage.emptyNote")}
              </p>
            )}
      </Card>
    </div>
  );
}
