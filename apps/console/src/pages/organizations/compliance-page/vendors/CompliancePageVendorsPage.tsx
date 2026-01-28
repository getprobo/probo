import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageVendorsPageQuery } from "#/__generated__/core/CompliancePageVendorsPageQuery.graphql";

import { CompliancePageVendorList } from "./_components/CompliancePageVendorList";

export const compliancePageVendorsPageQuery = graphql`
  query CompliancePageVendorsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ...CompliancePageVendorListFragment
    }
  }
`;

export function CompliancePageVendorsPage(props: {
  queryRef: PreloadedQuery<CompliancePageVendorsPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageVendorsPageQuery>(
    compliancePageVendorsPageQuery,
    queryRef,
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Vendors")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage vendor assessments and third-party risk information")}
          </p>
        </div>
      </div>

      <CompliancePageVendorList fragmentRef={organization} />
    </div>
  );
}
