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
