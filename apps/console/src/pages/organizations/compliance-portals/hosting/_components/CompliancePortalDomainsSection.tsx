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

import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDomainsSection_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalDomainsSection_compliancePortalFragment.graphql";
import type { CompliancePortalDomainsSection_organizationFragment$key } from "#/__generated__/core/CompliancePortalDomainsSection_organizationFragment.graphql";

import { domainsSection } from "../variants";

import { CompliancePortalCustomDomainForm } from "./CompliancePortalCustomDomainForm";
import { CompliancePortalDomainCard } from "./CompliancePortalDomainCard";

const organizationFragment = graphql`
  fragment CompliancePortalDomainsSection_organizationFragment on Organization {
    canCreateCustomDomain: permission(action: "compliance-portal:custom-domain:create")
  }
`;

const compliancePortalFragment = graphql`
  fragment CompliancePortalDomainsSection_compliancePortalFragment on CompliancePortal {
    defaultDomain {
      ...CompliancePortalDomainCardFragment
    }
    customDomain {
      ...CompliancePortalDomainCardFragment
    }
  }
`;

interface CompliancePortalDomainsSectionProps {
  organizationKey: CompliancePortalDomainsSection_organizationFragment$key;
  compliancePortalKey: CompliancePortalDomainsSection_compliancePortalFragment$key;
}

export function CompliancePortalDomainsSection({
  organizationKey,
  compliancePortalKey,
}: CompliancePortalDomainsSectionProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, intro, grid } = domainsSection();

  const organization = useFragment(organizationFragment, organizationKey);
  const compliancePortal = useFragment(compliancePortalFragment, compliancePortalKey);
  const defaultDomain = compliancePortal.defaultDomain;
  const customDomain = compliancePortal.customDomain;
  const showCustomDomainForm = customDomain == null && organization.canCreateCustomDomain;
  const hasCards = defaultDomain != null || customDomain != null || showCustomDomainForm;

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("domainsSection.title")}
        </Heading>
        <Text size={2} color="neutral">
          {t("domainsSection.description")}
        </Text>
      </div>

      {hasCards && (
        <div className={grid()}>
          {defaultDomain && (
            <CompliancePortalDomainCard customDomainKey={defaultDomain} />
          )}
          {customDomain && (
            <CompliancePortalDomainCard customDomainKey={customDomain} />
          )}
          {showCustomDomainForm && <CompliancePortalCustomDomainForm />}
        </div>
      )}
    </section>
  );
}
