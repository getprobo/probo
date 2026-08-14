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

import { CheckIcon, ClockIcon, LockSimpleIcon } from "@phosphor-icons/react";
import { parseDate } from "@probo/helpers";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { OrganizationSwitcherMenuItem_membership$key } from "#/__generated__/iam/OrganizationSwitcherMenuItem_membership.graphql";
import type { OrganizationSwitcherMenuItem_organization$key } from "#/__generated__/iam/OrganizationSwitcherMenuItem_organization.graphql";

const membershipFragment = graphql`
  fragment OrganizationSwitcherMenuItem_membership on Membership {
    lastSession {
      id
      expiresAt
    }
  }
`;

const organizationFragment = graphql`
  fragment OrganizationSwitcherMenuItem_organization on Organization {
    id
    name
    logo {
      downloadUrl
    }
  }
`;

const SESSION_STATES = {
  active: { Icon: CheckIcon, className: "text-green-11", labelKey: "organizationSwitcher.sessionActive" },
  expired: { Icon: ClockIcon, className: "text-amber-11", labelKey: "organizationSwitcher.sessionExpired" },
  locked: { Icon: LockSimpleIcon, className: "text-sand-11", labelKey: "organizationSwitcher.sessionLocked" },
} as const;

export interface OrganizationSwitcherMenuItemProps {
  membershipKey: OrganizationSwitcherMenuItem_membership$key;
  organizationKey: OrganizationSwitcherMenuItem_organization$key;
}

export function OrganizationSwitcherMenuItem(props: OrganizationSwitcherMenuItemProps) {
  const { membershipKey, organizationKey } = props;

  const { t } = useTranslation();
  const { lastSession } = useFragment(membershipFragment, membershipKey);
  const organization = useFragment(organizationFragment, organizationKey);

  let sessionState: keyof typeof SESSION_STATES = "locked";
  if (lastSession != null) {
    sessionState = parseDate(lastSession.expiresAt) < new Date() ? "expired" : "active";
  }
  const { Icon: SessionIcon, className, labelKey } = SESSION_STATES[sessionState];

  return (
    <DropdownItem
      iconStart={(
        <Avatar
          size={1}
          variant="soft"
          color="neutral"
          radius="small"
          src={organization.logo?.downloadUrl ?? undefined}
          fallback={organization.name.charAt(0) || "?"}
        />
      )}
      shortcut={<SessionIcon className={`size-4 ${className}`} aria-label={t(labelKey)} />}
      render={<Link to={`/organizations/${organization.id}`} />}
    >
      {organization.name}
    </DropdownItem>
  );
}
