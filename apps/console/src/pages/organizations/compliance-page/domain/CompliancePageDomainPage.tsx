import { useTranslate } from "@probo/i18n";
import { Button, Card, IconPlusLarge } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDomainPageQuery } from "#/__generated__/core/CompliancePageDomainPageQuery.graphql";

import { CompliancePageDomainCard } from "./_components/CompliancePageDomainCard";
import { NewCompliancePageDomainDialog } from "./_components/NewCompliancePageDomainDialog";

export const compliancePageDomainPageQuery = graphql`
  query CompliancePageDomainPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateCustomDomain: permission(action: "core:custom-domain:create")
        customDomain {
          ...CompliancePageDomainCardFragment
        }
      }
    }
  }
`;

export function CompliancePageDomainPage(props: {
  queryRef: PreloadedQuery<CompliancePageDomainPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageDomainPageQuery>(
    compliancePageDomainPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Custom Domain")}</h2>
      {organization.customDomain
        ? (
            <CompliancePageDomainCard fKey={organization.customDomain} />
          )
        : (
            <Card padded>
              <div className="text-center py-8">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No custom domain configured")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__(
                    "Add your own domain to make your compliance page more professional",
                  )}
                </p>
                <div className="flex justify-center">
                  {organization.canCreateCustomDomain && (
                    <NewCompliancePageDomainDialog>
                      <Button icon={IconPlusLarge}>{__("Add Domain")}</Button>
                    </NewCompliancePageDomainDialog>
                  )}
                </div>
              </div>
            </Card>
          )}
    </div>
  );
}
