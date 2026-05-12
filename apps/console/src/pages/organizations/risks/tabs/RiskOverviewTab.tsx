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

import { RiskOverview } from "@probo/ui";
import { useFragment } from "react-relay";
import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";

import type { RiskOverviewTabFragment$key } from "#/__generated__/core/RiskOverviewTabFragment.graphql";

const overviewFragment = graphql`
  fragment RiskOverviewTabFragment on Risk {
    # eslint-disable-next-line relay/unused-fields
    inherentLikelihood
    # eslint-disable-next-line relay/unused-fields
    inherentImpact
    # eslint-disable-next-line relay/unused-fields
    residualLikelihood
    # eslint-disable-next-line relay/unused-fields
    residualImpact
  }
`;

export default function RiskOverviewTab() {
  const { risk: key } = useOutletContext<{
    risk: RiskOverviewTabFragment$key;
  }>();

  const risk = useFragment(overviewFragment, key);
  return (
    <div className="grid grid-cols-2 gap-4">
      <RiskOverview type="inherent" risk={risk} />
      <RiskOverview type="residual" risk={risk} />
    </div>
  );
}
