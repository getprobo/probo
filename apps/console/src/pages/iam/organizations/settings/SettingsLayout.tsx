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

import { useTranslate } from "@probo/i18n";
import {
  IconKey,
  IconListStack,
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
        <TabLink to={`/organizations/${organizationId}/settings/audit-log`}>
          <IconListStack size={20} />
          {__("Audit Log")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
