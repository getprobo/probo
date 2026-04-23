// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Button, Card, IconPlusLarge, useToast } from "@probo/ui";
import { useEffect } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";
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
  const { toast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();

  useEffect(() => {
    const dcResult = searchParams.get("domain_connect");
    if (!dcResult) return;

    if (dcResult === "success") {
      toast({
        title: __("DNS configured"),
        description: __("DNS record created successfully via Domain Connect"),
        variant: "success",
      });
    } else if (dcResult === "error") {
      toast({
        title: __("Configuration failed"),
        description: searchParams.get("error_description") ?? __("DNS configuration via Domain Connect failed"),
        variant: "error",
      });
    }

    setSearchParams((prev) => {
      prev.delete("domain_connect");
      prev.delete("error_description");
      return prev;
    }, { replace: true });
  }, [searchParams, setSearchParams, toast, __]);

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
