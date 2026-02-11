import { useTranslate } from "@probo/i18n";
import {
  IconKey,
  IconLock,
  IconSend,
  IconSettingsGear2,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { Outlet } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";

export default function SettingsLayout() {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <div className="space-y-6">
      <PageHeader title={__("Settings")} />

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/settings/general`}>
          <IconSettingsGear2 size={20} />
          {__("General")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/settings/saml-sso`}>
          <IconLock size={20} />
          {__("SAML SSO")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/settings/scim`}>
          <IconKey size={20} />
          {__("SCIM")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/settings/webhooks`}>
          <IconSend size={20} />
          {__("Webhooks")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
