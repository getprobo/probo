import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  IconCheckmark1,
  IconFolder2,
  IconMedal,
  IconPageTextLine,
  IconPeopleAdd,
  IconSettingsGear2,
  IconStore,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import type { TrustCenterGraphQuery } from "#/__generated__/core/TrustCenterGraphQuery.graphql";
import { trustCenterQuery } from "#/hooks/graph/TrustCenterGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  queryRef: PreloadedQuery<TrustCenterGraphQuery>;
};

export default function TrustCenterPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { organization } = usePreloadedQuery(trustCenterQuery, queryRef);

  usePageTitle(__("Trust Center"));

  const isActive = organization.trustCenter?.active || false;

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Trust Center")}
        description={__(
          "Configure your public trust center to showcase your security and compliance posture.",
        )}
      >
        <Badge variant={isActive ? "success" : "danger"}>
          {isActive ? __("Active") : __("Inactive")}
        </Badge>
      </PageHeader>

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/trust-center`} end>
          <IconSettingsGear2 className="size-4" />
          {__("Overview")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/trust-center/trust-by`}>
          <IconCheckmark1 className="size-4" />
          {__("Trusted by")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/trust-center/audits`}>
          <IconMedal className="size-4" />
          {__("Audits")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/trust-center/documents`}>
          <IconPageTextLine className="size-4" />
          {__("Documents")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/trust-center/files`}>
          <IconFolder2 className="size-4" />
          {__("Files")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/trust-center/vendors`}>
          <IconStore className="size-4" />
          {__("Vendors")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/trust-center/access`}>
          <IconPeopleAdd className="size-4" />
          {__("Access")}
        </TabLink>
      </Tabs>

      <Outlet context={{ organization }} />
    </div>
  );
}
