// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
