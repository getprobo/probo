import { useFragment } from "react-relay";
import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";
import type { RiskOverviewTabFragment$key } from "./__generated__/RiskOverviewTabFragment.graphql";
import { RiskOverview } from "@probo/ui";

const overviewFragment = graphql`
  fragment RiskOverviewTabFragment on Risk {
    inherentLikelihood
    inherentImpact
    residualLikelihood
    residualImpact
    inherentRiskScore
    residualRiskScore
  }
`;

export default function RiskOverviewTab() {
  const { risk: key } = useOutletContext<{
    risk: RiskOverviewTabFragment$key;
  }>();

  const risk = useFragment(overviewFragment, key);
  return (
    <div className="grid grid-cols-2">
      <RiskOverview type="inherent" risk={risk} />
      <RiskOverview type="residual" risk={risk} />
    </div>
  );
}
