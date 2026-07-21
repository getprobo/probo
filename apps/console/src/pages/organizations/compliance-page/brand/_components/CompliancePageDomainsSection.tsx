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

import { Button, IconChevronRight } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDomainsSection_organizationFragment$key } from "#/__generated__/core/CompliancePageDomainsSection_organizationFragment.graphql";

import { CompliancePageDomainCard } from "../../domain/_components/CompliancePageDomainCard";
import { NewCompliancePageDomainDialog } from "../../domain/_components/NewCompliancePageDomainDialog";

const organizationFragment = graphql`
  fragment CompliancePageDomainsSection_organizationFragment on Organization {
    canCreateCustomDomain: permission(action: "compliance-portal:custom-domain:create")
    compliancePage: compliancePortal @required(action: THROW) {
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
  const { t } = useTranslation("organizations/compliance-page");

  const organization = useFragment(organizationFragment, props.organizationRef);
  const compliancePageId = organization.compliancePage.id;
  const defaultDomain = organization.compliancePage.defaultDomain;
  const customDomain = organization.compliancePage.customDomain;

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{t("brandPage.domains.title")}</h2>
        <p className="text-sm text-txt-tertiary">
          {t("brandPage.domains.description")}
        </p>
      </div>

      <div className="space-y-3">
        {defaultDomain && (
          <CompliancePageDomainCard fKey={defaultDomain} />
        )}

        {customDomain
          ? (
              <CompliancePageDomainCard fKey={customDomain} />
            )
          : organization.canCreateCustomDomain && (
            <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border-solid px-4 py-8">
              <p className="max-w-md text-center text-sm text-txt-tertiary">
                {t("domainPage.empty.description")}
              </p>
              <NewCompliancePageDomainDialog compliancePageId={compliancePageId}>
                <Button iconAfter={IconChevronRight}>{t("brandPage.domains.actions.configure")}</Button>
              </NewCompliancePageDomainDialog>
            </div>
          )}
      </div>
    </section>
  );
}
