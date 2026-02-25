import { useTranslate } from "@probo/i18n";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageNewsletterPageQuery } from "#/__generated__/core/CompliancePageNewsletterPageQuery.graphql";

import { CompliancePageNewsletterList } from "./_components/CompliancePageNewsletterList";

export const compliancePageNewsletterPageQuery = graphql`
  query CompliancePageNewsletterPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter @required(action: THROW) {
          id
          ...CompliancePageNewsletterListFragment
        }
      }
    }
  }
`;

export function CompliancePageNewsletterPage(props: {
  queryRef: PreloadedQuery<CompliancePageNewsletterPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageNewsletterPageQuery>(
    compliancePageNewsletterPageQuery,
    queryRef,
  );

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-base font-medium">{__("Newsletter Subscribers")}</h3>
        <p className="text-sm text-txt-tertiary">
          {__("People subscribed to receive security and compliance updates")}
        </p>
      </div>
      <CompliancePageNewsletterList fragmentRef={organization.compliancePage} />
    </div>
  );
}
