import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageBrandPageQuery } from "#/__generated__/core/CompliancePageBrandPageQuery.graphql";

export const compliancePageBrandPageQuery = graphql`
  query CompliancePageBrandPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        compliancePage: trustCenter {
          logoFileUrl
          darkLogoFileUrl
        }
      }
    }
  }
`;

export function CompliancePageBrandPage(props: { queryRef: PreloadedQuery<CompliancePageBrandPageQuery> }) {
  const { queryRef } = props;

  const { organization } = usePreloadedQuery<CompliancePageBrandPageQuery>(compliancePageBrandPageQuery, queryRef);

  return <pre>{JSON.stringify(organization, null, 2)}</pre>;
}
