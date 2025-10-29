import { Outlet } from "react-router";
import { useTranslate } from "@probo/i18n";
import type { PreloadedQuery } from "react-relay";
import type { OrganizationGraph_ViewQuery } from "/hooks/graph/__generated__/OrganizationGraph_ViewQuery.graphql";
import { usePreloadedQuery } from "react-relay";
import { organizationViewQuery } from "/hooks/graph/OrganizationGraph";
import {
  IconSettingsGear2,
  IconPeopleAdd,
  IconStore,
  IconLock,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { graphql } from "relay-runtime";
import type { SettingsPageFragment$key } from "./__generated__/SettingsPageFragment.graphql";
import { useFragment } from "react-relay";

const organizationFragment = graphql`
  fragment SettingsPageFragment on Organization {
    id
    name
    ...GeneralSettingsTabFragment
    ...MembersSettingsTabMembershipsFragment
    ...MembersSettingsTabInvitationsFragment
    ...DomainSettingsTabFragment
    ...SAMLSettingsTabFragment
  }
`;

type Props = {
  queryRef: PreloadedQuery<OrganizationGraph_ViewQuery>;
};

export default function SettingsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const organizationKey = usePreloadedQuery(
    organizationViewQuery,
    queryRef
  ).node;
  const organization = useFragment<SettingsPageFragment$key>(
    organizationFragment,
    organizationKey
  );

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

      <Outlet context={{ organization }} />
    </div>
  );
}
