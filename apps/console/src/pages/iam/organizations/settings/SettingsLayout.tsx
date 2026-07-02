// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  IconGroup1,
  IconKey,
  IconListStack,
  IconLock,
  IconSend,
  IconSettingsGear2,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { SettingsLayoutFragment$key } from "#/__generated__/iam/SettingsLayoutFragment.graphql";
import type { SettingsLayoutQuery } from "#/__generated__/iam/SettingsLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const settingsLayoutQuery = graphql`
  query SettingsLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...SettingsLayoutFragment
      }
    }
  }
`;

const fragment = graphql`
  fragment SettingsLayoutFragment on Organization {
    canListMembers: permission(action: "iam:membership:list")
    canUpdateOrganization: permission(action: "iam:organization:update")
  }
`;

export function SettingsLayout(props: {
  queryRef: PreloadedQuery<SettingsLayoutQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<SettingsLayoutQuery>(
    settingsLayoutQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  const permissions = useFragment<SettingsLayoutFragment$key>(
    fragment,
    organization,
  );

  const prefix = `/organizations/${organizationId}/settings`;

  return (
    <div className="space-y-6">
      <PageHeader title={__("Organization")} />

      <Tabs>
        {permissions.canListMembers && (
          <TabLink to={`${prefix}/members`}>
            <IconGroup1 size={20} />
            {__("Members")}
          </TabLink>
        )}
        {permissions.canUpdateOrganization && (
          <>
            <TabLink to={`${prefix}/general`}>
              <IconSettingsGear2 size={20} />
              {__("General")}
            </TabLink>
            <TabLink to={`${prefix}/saml-sso`}>
              <IconLock size={20} />
              {__("SAML SSO")}
            </TabLink>
            <TabLink to={`${prefix}/scim`}>
              <IconKey size={20} />
              {__("SCIM")}
            </TabLink>
            <TabLink to={`${prefix}/webhooks`}>
              <IconSend size={20} />
              {__("Webhooks")}
            </TabLink>
            <TabLink to={`${prefix}/audit-log`}>
              <IconListStack size={20} />
              {__("Audit Log")}
            </TabLink>
          </>
        )}
      </Tabs>

      <Outlet />
    </div>
  );
}
