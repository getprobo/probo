import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageDocumentsPageQuery } from "#/__generated__/core/CompliancePageDocumentsPageQuery.graphql";

import { CompliancePageDocumentList } from "./_components/CompliancePageDocumentList";

export const compliancePageDocumentsPageQuery = graphql`
  query CompliancePageDocumentsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ...CompliancePageDocumentListFragment
    }
  }
`;

export function CompliancePageDocumentsPage(props: { queryRef: PreloadedQuery<CompliancePageDocumentsPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageDocumentsPageQuery>(
    compliancePageDocumentsPageQuery,
    queryRef,
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Documents")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage policies, procedures and compliance documents")}
          </p>
        </div>
      </div>

      <CompliancePageDocumentList fragmentRef={organization} />
    </div>
  );
};
