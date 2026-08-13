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

import { PageHeader, TabLink, Tabs } from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet } from "react-router";

import type { EmployeeTabsLayoutQuery } from "#/__generated__/core/EmployeeTabsLayoutQuery.graphql";

export const employeeTabsLayoutQuery = graphql`
  query EmployeeTabsLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        slackbotInstallation {
          active
        }
      }
    }
  }
`;

interface EmployeeTabsLayoutProps {
  queryRef: PreloadedQuery<EmployeeTabsLayoutQuery>;
}

export function EmployeeTabsLayout({ queryRef }: EmployeeTabsLayoutProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<EmployeeTabsLayoutQuery>(
    employeeTabsLayoutQuery,
    queryRef,
  );

  if (data.organization.__typename !== "Organization") {
    throw new Error("Relay node is not an organization");
  }

  const showSlackTab = data.organization.slackbotInstallation?.active === true;

  return (
    <div className="space-y-6">
      <PageHeader title={t("employeeTabsLayout.title")} />
      <Tabs>
        <TabLink to="signatures" end>
          {t("employeeTabsLayout.tabs.signatures")}
        </TabLink>
        <TabLink to="approvals" end>
          {t("employeeTabsLayout.tabs.approvals")}
        </TabLink>
        <TabLink to="devices" end>
          {t("devices.title")}
        </TabLink>
        {showSlackTab && (
          <TabLink to="bindings" end>
            {t("employeeTabsLayout.tabs.bindings", { defaultValue: "Slack" })}
          </TabLink>
        )}
      </Tabs>
      <Outlet />
    </div>
  );
}
