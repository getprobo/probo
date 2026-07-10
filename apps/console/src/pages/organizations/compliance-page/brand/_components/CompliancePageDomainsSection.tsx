// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { Button, IconChevronRight } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDomainsSection_organizationFragment$key } from "#/__generated__/core/CompliancePageDomainsSection_organizationFragment.graphql";

import { CompliancePageDomainCard } from "../../domain/_components/CompliancePageDomainCard";
import { NewCompliancePageDomainDialog } from "../../domain/_components/NewCompliancePageDomainDialog";

const organizationFragment = graphql`
  fragment CompliancePageDomainsSection_organizationFragment on Organization {
    canCreateCustomDomain: permission(action: "compliance-portal:custom-domain:create")
    compliancePage: trustCenter @required(action: THROW) {
      id
      defaultDomain {
        id
        ...CompliancePageDomainCardFragment
      }
      customDomain {
        id
        ...CompliancePageDomainCardFragment
      }
    }
  }
`;

export function CompliancePageDomainsSection(props: {
  organizationRef: CompliancePageDomainsSection_organizationFragment$key;
}) {
  const { __ } = useTranslate();

  const organization = useFragment(organizationFragment, props.organizationRef);
  const compliancePageId = organization.compliancePage.id;
  const defaultDomain = organization.compliancePage.defaultDomain;
  const customDomain = organization.compliancePage.customDomain;

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{__("Domains")}</h2>
        <p className="text-sm text-txt-tertiary">
          {__(
            "Your compliance page is always available on its default probopage.com subdomain. You can also serve it on one custom domain of your own.",
          )}
        </p>
      </div>

      <div className="space-y-3">
        {defaultDomain && (
          <CompliancePageDomainCard
            fKey={defaultDomain}
            compliancePageId={compliancePageId}
          />
        )}

        {customDomain
          ? (
              <CompliancePageDomainCard
                fKey={customDomain}
                compliancePageId={compliancePageId}
              />
            )
          : organization.canCreateCustomDomain && (
              <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border-solid px-4 py-8">
                <p className="max-w-md text-center text-sm text-txt-tertiary">
                  {__(
                    "Use your own domain to make your compliance page feel more professional.",
                  )}
                </p>
                <NewCompliancePageDomainDialog compliancePageId={compliancePageId}>
                  <Button iconAfter={IconChevronRight}>{__("Configure")}</Button>
                </NewCompliancePageDomainDialog>
              </div>
            )}
      </div>
    </section>
  );
}
