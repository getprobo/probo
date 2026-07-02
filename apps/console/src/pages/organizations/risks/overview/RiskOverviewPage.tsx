// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { RiskOverview } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { RiskOverviewPageQuery } from "#/__generated__/core/RiskOverviewPageQuery.graphql";

export const riskOverviewPageQuery = graphql`
  query RiskOverviewPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        inherentLikelihood
        inherentImpact
        residualLikelihood
        residualImpact
      }
    }
  }
`;

interface RiskOverviewPageProps {
  queryRef: PreloadedQuery<RiskOverviewPageQuery>;
}

export default function RiskOverviewPage(props: RiskOverviewPageProps) {
  const data = usePreloadedQuery<RiskOverviewPageQuery>(riskOverviewPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const { inherentLikelihood, inherentImpact, residualLikelihood, residualImpact }
    = data.node;
  const risk = {
    inherentLikelihood,
    inherentImpact,
    residualLikelihood,
    residualImpact,
  };
  return (
    <div className="grid grid-cols-2 gap-4">
      <RiskOverview type="inherent" risk={risk} />
      <RiskOverview type="residual" risk={risk} />
    </div>
  );
}
