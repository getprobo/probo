import {
  IconLock,
  IconPeopleAdd,
  IconSettingsGear2,
  IconStore,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { Outlet } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useTranslate } from "@probo/i18n";

export default function () {
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
        <TabLink to={`/organizations/${organizationId}/settings/members`}>
          <IconPeopleAdd size={20} />
          {__("Members")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/settings/domain`}>
          <IconStore size={20} />
          {__("Domain")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/settings/saml-sso`}>
          <IconLock size={20} />
          {__("SAML SSO")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
