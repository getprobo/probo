import { useOutletContext } from "react-router";
import { useFragment, graphql } from "react-relay";
import { useTranslate } from "@probo/i18n";
import { CustomDomainManager } from "/components/customDomains/CustomDomainManager";
import type { DomainSettingsTabFragment$key } from "./__generated__/DomainSettingsTabFragment.graphql";

const domainSettingsTabFragment = graphql`
  fragment DomainSettingsTabFragment on Organization {
    id
    customDomain {
      id
      domain
      sslStatus
      dnsRecords {
        type
        name
        value
        ttl
        purpose
      }
      createdAt
      updatedAt
      sslExpiresAt
    }
  }
`;

type OutletContext = {
  organization: DomainSettingsTabFragment$key;
};

export default function DomainSettingsTab() {
  const { __ } = useTranslate();
  const { organization: organizationKey } = useOutletContext<OutletContext>();
  const organization = useFragment(domainSettingsTabFragment, organizationKey);

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Custom Domain")}</h2>
      <CustomDomainManager
        organizationId={organization.id}
        customDomain={organization.customDomain}
      />
    </div>
  );
}
