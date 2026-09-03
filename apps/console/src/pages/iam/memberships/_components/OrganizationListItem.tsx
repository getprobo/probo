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

import { CheckIcon, ClockIcon, LockIcon } from "@phosphor-icons/react";
import { getMembershipSessionStatus } from "@probo/helpers";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { OrganizationListItem_profile$key } from "#/__generated__/iam/OrganizationListItem_profile.graphql";

const organizationListItemFragment = graphql`
  fragment OrganizationListItem_profile on Profile {
    state
    membership @required(action: THROW) {
      lastSession {
        expiresAt
      }
    }
    organization @required(action: THROW) {
      id
      name
      logo {
        downloadUrl
      }
    }
  }
`;

interface OrganizationListItemProps {
  profileKey: OrganizationListItem_profile$key;
}

export function OrganizationListItem({ profileKey }: OrganizationListItemProps) {
  const { t } = useTranslation();
  const { membership, organization, state } = useFragment(
    organizationListItemFragment,
    profileKey,
  );

  const sessionStatus = getMembershipSessionStatus(
    membership.lastSession == null
      ? null
      : { expiresAt: membership.lastSession.expiresAt },
  );
  const isAssuming = sessionStatus === "authenticated";

  let statusBadge = (
    <Badge color="neutral" variant="soft" iconStart={<LockIcon />}>
      {t("membershipCard.status.authenticationRequired")}
    </Badge>
  );
  if (sessionStatus === "authenticated") {
    statusBadge = (
      <Badge color="green" variant="soft" iconStart={<CheckIcon />}>
        {t("membershipCard.status.authenticated")}
      </Badge>
    );
  } else if (sessionStatus === "expired") {
    statusBadge = (
      <Badge color="amber" variant="soft" iconStart={<ClockIcon />}>
        {t("membershipCard.status.sessionExpired")}
      </Badge>
    );
  }

  return (
    <ListItem>
      <Avatar
        size={4}
        variant="soft"
        color="neutral"
        radius="small"
        src={organization.logo?.downloadUrl ?? undefined}
        fallback={organization.name.charAt(0) || "?"}
      />
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast>
          {organization.name}
        </Text>
        {statusBadge}
      </ListItemContent>
      {state === "ACTIVE"
        ? (
            <ButtonLink
              to={`/organizations/${organization.id}`}
              size={2}
              variant={isAssuming ? "soft" : "solid"}
              color={isAssuming ? "neutral" : "gold"}
            >
              {isAssuming
                ? t("membershipCard.actions.continue")
                : t("membershipCard.actions.open")}
            </ButtonLink>
          )
        : (
            <Button size={2} variant="soft" color="neutral" disabled>
              {t("membershipCard.accountDeactivated")}
            </Button>
          )}
    </ListItem>
  );
}
