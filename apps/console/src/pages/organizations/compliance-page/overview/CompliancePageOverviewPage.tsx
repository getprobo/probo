import { useTranslate } from "@probo/i18n";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageOverviewPageQuery } from "#/__generated__/core/CompliancePageOverviewPageQuery.graphql";

import { CompliancePageFrameworkList } from "./_components/CompliancePageFrameworkList";
import { CompliancePageNDASection } from "./_components/CompliancePageNDASection";
import { CompliancePageSlackSection } from "./_components/CompliancePageSlackSection";
import { CompliancePageStatusSection } from "./_components/CompliancePageStatusSection";

export const compliancePageOverviewPageQuery = graphql`
  query CompliancePageOverviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        compliancePage: trustCenter {
          canGetNDA: permission(action: "core:trust-center:get-nda")
          ...CompliancePageFrameworkList_compliancePageFragment
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

  const { __ } = useTranslate();

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

      {organization.compliancePage && (
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-medium">{__("Frameworks")}</h3>
            <p className="text-sm text-txt-tertiary">
              {__("Select which frameworks to display as badges on your compliance page")}
            </p>
          </div>
          <CompliancePageFrameworkList compliancePageRef={organization.compliancePage} />
        </div>
      )}
    </div>
  );
}
