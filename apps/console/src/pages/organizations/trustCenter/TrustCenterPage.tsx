import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  PageHeader,
  Tabs,
  TabLink,
  Badge,
  IconSettingsGear2,
  IconMedal,
  IconCheckmark1,
  IconPageTextLine,
  IconFolder2,
  IconStore,
  IconPeopleAdd,
} from "@probo/ui";
import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { trustCenterQuery } from "/hooks/graph/TrustCenterGraph";
import type { TrustCenterGraphQuery } from "/hooks/graph/__generated__/TrustCenterGraphQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Outlet } from "react-router";

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
          "Configure your public trust center to showcase your security and compliance posture."
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
