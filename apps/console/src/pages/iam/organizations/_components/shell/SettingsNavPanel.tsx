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

import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { SettingsNavPanelQuery } from "#/__generated__/iam/SettingsNavPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";

import { NavPanelGroup } from "./NavPanelGroup";
import { NavPanelItem } from "./NavPanelItem";
import { NavPanelQuery } from "./NavPanelQuery";
import type { NavPanelBodyProps } from "./navPanels";

const settingsNavPanelQuery = graphql`
  query SettingsNavPanelQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canUpdateOrganization: permission(action: "iam:organization:update")
        canGetContext: permission(action: "core:organization-context:get")
        canListWebhookSubscriptions: permission(action: "core:webhook-subscription:list")
        canConnectSlack: permission(action: "core:connector:initiate")
        canUninstallSlack: permission(action: "core:connector:delete")
        canListMembers: permission(action: "iam:membership:list")
        canListAuditLogEntries: permission(action: "iam:audit-log-entry:list")
      }
    }
  }
`;

export function SettingsNavPanel({ group }: NavPanelBodyProps) {
  return (
    <NavPanelQuery<SettingsNavPanelQuery> query={settingsNavPanelQuery}>
      {queryRef => <SettingsNavPanelInner queryRef={queryRef} group={group} />}
    </NavPanelQuery>
  );
}

interface SettingsNavPanelInnerProps extends NavPanelBodyProps {
  queryRef: PreloadedQuery<SettingsNavPanelQuery>;
}

function SettingsNavPanelInner({ queryRef, group }: SettingsNavPanelInnerProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<SettingsNavPanelQuery>(settingsNavPanelQuery, queryRef);
  const { organization } = data;
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }
  const showIam = organization.canListMembers
    || organization.canUpdateOrganization
    || organization.canListAuditLogEntries;

  return (
    <>
      {organization.canUpdateOrganization && (
        <NavPanelItem
          label={t("nav.organization")}
          to={navHref(organizationId, group, "general")}
        />
      )}
      {organization.canGetContext && (
        <NavPanelItem
          label={t("nav.context")}
          to={navHref(organizationId, group, "context")}
        />
      )}
      {organization.canListWebhookSubscriptions && (
        <NavPanelItem
          label={t("nav.webhooks")}
          to={navHref(organizationId, group, "webhooks")}
        />
      )}
      {(organization.canConnectSlack || organization.canUninstallSlack) && (
        <NavPanelItem
          label={t("nav.slackBot")}
          to={navHref(organizationId, group, "slackbot")}
        />
      )}
      {showIam && (
        <NavPanelGroup label={t("nav.iam")}>
          {organization.canListMembers && (
            <NavPanelItem
              label={t("nav.users")}
              to={navHref(organizationId, group, "people")}
            />
          )}
          {organization.canUpdateOrganization && (
            <NavPanelItem
              label={t("nav.authProvisioning")}
              to={navHref(organizationId, group, "auth")}
            />
          )}
          {organization.canListAuditLogEntries && (
            <NavPanelItem
              label={t("nav.auditLog")}
              to={navHref(organizationId, group, "audit-log")}
            />
          )}
        </NavPanelGroup>
      )}
    </>
  );
}
