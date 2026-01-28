import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageOverviewPageQuery } from "#/__generated__/core/CompliancePageOverviewPageQuery.graphql";

import { CompliancePageNDASection } from "./_components/CompliancePageNDASection";
import { CompliancePageSlackSection } from "./_components/CompliancePageSlackSection";
import { CompliancePageStatusSection } from "./_components/CompliancePageStatusSection";

export const compliancePageOverviewPageQuery = graphql`
  query CompliancePageOverviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        compliancePage: trustCenter {
          canGetNDA: permission(action: "core:trust-center:get-nda")
        }
      }
      ...CompliancePageStatusSectionFragment
      ...CompliancePageNDASectionFragment
      ...CompliancePageSlackSectionFragment
    }
  }
`;

export function CompliancePageOverviewPage(props: { queryRef: PreloadedQuery<CompliancePageOverviewPageQuery> }) {
  const { queryRef } = props;

  const { organization } = usePreloadedQuery<CompliancePageOverviewPageQuery>(
    compliancePageOverviewPageQuery,
    queryRef,
  );

  return (
    <div className="space-y-6">
      <CompliancePageStatusSection fragmentRef={organization} />

      {organization.compliancePage?.canGetNDA && (
        <CompliancePageNDASection fragmentRef={organization} />
      )}

      <CompliancePageSlackSection fragmentRef={organization} />
    </div>
  );
}
