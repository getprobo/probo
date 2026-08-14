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

import type { CompliancePortalDomainsSection_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalDomainsSection_compliancePortalFragment.graphql";
import type { CompliancePortalDomainsSection_organizationFragment$key } from "#/__generated__/core/CompliancePortalDomainsSection_organizationFragment.graphql";

import { CompliancePortalDomainCard } from "./CompliancePortalDomainCard";
import { NewCompliancePortalDomainDialog } from "./NewCompliancePortalDomainDialog";

const organizationFragment = graphql`
  fragment CompliancePortalDomainsSection_organizationFragment on Organization {
    canCreateCustomDomain: permission(action: "compliance-portal:custom-domain:create")
  }
`;

const compliancePortalFragment = graphql`
  fragment CompliancePortalDomainsSection_compliancePortalFragment on CompliancePortal {
    id
    defaultDomain {
      id
      ...CompliancePortalDomainCardFragment
    }
    customDomain {
      id
      ...CompliancePortalDomainCardFragment
    }
  }
`;

export function CompliancePortalDomainsSection(props: {
  organizationRef: CompliancePortalDomainsSection_organizationFragment$key;
  compliancePortalRef: CompliancePortalDomainsSection_compliancePortalFragment$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");

  const organization = useFragment(organizationFragment, props.organizationRef);
  const compliancePortal = useFragment(compliancePortalFragment, props.compliancePortalRef);
  const compliancePortalId = compliancePortal.id;
  const defaultDomain = compliancePortal.defaultDomain;
  const customDomain = compliancePortal.customDomain;

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
          <CompliancePortalDomainCard fKey={defaultDomain} />
        )}

        {customDomain
          ? (
              <CompliancePortalDomainCard fKey={customDomain} />
            )
          : organization.canCreateCustomDomain && (
            <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border-solid px-4 py-8">
              <p className="max-w-md text-center text-sm text-txt-tertiary">
                {t("domainPage.empty.description")}
              </p>
              <NewCompliancePortalDomainDialog compliancePortalId={compliancePortalId}>
                <Button iconAfter={IconChevronRight}>{t("brandPage.domains.actions.configure")}</Button>
              </NewCompliancePortalDomainDialog>
            </div>
          )}
      </div>
    </section>
  );
}
