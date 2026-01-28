import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageAuditsPageQuery } from "#/__generated__/core/CompliancePageAuditsPageQuery.graphql";

import { CompliancePageAuditList } from "./_components/CompliancePageAuditList";

export const compliancePageAuditsPageQuery = graphql`
  query CompliancePageAuditsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ...CompliancePageAuditListFragment
    }
  }
`;

export function CompliancePageAuditsPage(props: { queryRef: PreloadedQuery<CompliancePageAuditsPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageAuditsPageQuery>(compliancePageAuditsPageQuery, queryRef);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Audits")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage audit reports and compliance certifications")}
          </p>
        </div>
      </div>

      <CompliancePageAuditList fragmentRef={organization} />
    </div>
  );
}
