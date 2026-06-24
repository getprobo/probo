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
