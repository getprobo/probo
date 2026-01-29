import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Badge, IconCheckmark1, IconFolder2, IconMedal, IconPageTextLine, IconPencil, IconPeopleAdd, IconSettingsGear2, IconStore, PageHeader, TabLink, Tabs } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePageLayoutQuery } from "#/__generated__/core/CompliancePageLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const compliancePageLayoutQuery = graphql`
  query CompliancePageLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter {
          active
        }
      }
    }
  }
`;

export function CompliancePageLayout(props: { queryRef: PreloadedQuery<CompliancePageLayoutQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  usePageTitle(__("Compliance Page"));

  const { organization } = usePreloadedQuery<CompliancePageLayoutQuery>(compliancePageLayoutQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Compliance Page")}
        description={__(
          "Configure your public compliance page to showcase your security and compliance posture.",
        )}
      >
        <Badge variant={organization.compliancePage?.active ? "success" : "danger"}>
          {organization.compliancePage?.active ? __("Active") : __("Inactive")}
        </Badge>
      </PageHeader>

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/compliance-page`} end>
          <IconSettingsGear2 className="size-4" />
          {__("Overview")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/brand`}>
          <IconPencil className="size-4" />
          {__("Brand")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/domain`}>
          <IconStore className="size-4" />
          {__("Domain")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/references`}>
          <IconCheckmark1 className="size-4" />
          {__("References")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/audits`}>
          <IconMedal className="size-4" />
          {__("Audits")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/documents`}>
          <IconPageTextLine className="size-4" />
          {__("Documents")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/files`}>
          <IconFolder2 className="size-4" />
          {__("Files")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/vendors`}>
          <IconStore className="size-4" />
          {__("Vendors")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/access`}>
          <IconPeopleAdd className="size-4" />
          {__("Access")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
