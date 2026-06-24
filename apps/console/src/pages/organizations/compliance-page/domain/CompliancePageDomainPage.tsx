// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
