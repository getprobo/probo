// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import {
  IconSend,
  IconSettingsGear2,
  PageHeader,
  Slack,
  TabLink,
  Tabs,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { SettingsLayoutQuery } from "#/__generated__/core/SettingsLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const settingsLayoutQuery = graphql`
  query SettingsLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        slackbotAvailable
        canConnectSlack: permission(action: "core:connector:initiate")
        canUninstallSlack: permission(action: "core:connector:delete")
      }
    }
  }
`;

interface SettingsLayoutProps {
  queryRef: PreloadedQuery<SettingsLayoutQuery>;
}

export function SettingsLayout({ queryRef }: SettingsLayoutProps) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const data = usePreloadedQuery<SettingsLayoutQuery>(
    settingsLayoutQuery,
    queryRef,
  );

  if (data.organization.__typename !== "Organization") {
    throw new Error("Relay node is not an organization");
  }

  const showSlackTab = data.organization.slackbotAvailable
    && (data.organization.canConnectSlack
      || data.organization.canUninstallSlack);

  return (
    <div className="space-y-6">
      <PageHeader title={t("settingsLayout.title")} />

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/settings/general`}>
          <IconSettingsGear2 size={20} />
          {t("settingsLayout.tabs.general")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/settings/webhooks`}>
          <IconSend size={20} />
          {t("settingsLayout.tabs.webhooks")}
        </TabLink>
        {showSlackTab && (
          <TabLink to={`/organizations/${organizationId}/settings/slackbot`}>
            <Slack className="h-5 w-5" />
            {t("settingsLayout.tabs.slackBot")}
          </TabLink>
        )}
      </Tabs>

      <Outlet />
    </div>
  );
}
