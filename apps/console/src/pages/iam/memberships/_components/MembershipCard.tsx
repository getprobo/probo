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

import { parseDate } from "@probo/helpers";
import {
  Avatar,
  Badge,
  Button,
  Card,
  IconCheckmark1,
  IconClock,
  IconLock,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { MembershipCard_organizationFragment$key } from "#/__generated__/iam/MembershipCard_organizationFragment.graphql";
import type { MembershipCardFragment$key } from "#/__generated__/iam/MembershipCardFragment.graphql";

const fragment = graphql`
  fragment MembershipCardFragment on Profile {
    state
    membership @required(action: THROW) {
      lastSession {
        id
        expiresAt
      }
    }
  }
`;

const organizationFragment = graphql`
  fragment MembershipCard_organizationFragment on Organization {
    id
    name
    logo {
      downloadUrl
    }
  }
`;

interface MembershipCardProps {
  fKey: MembershipCardFragment$key;
  organizationFragmentRef: MembershipCard_organizationFragment$key;
}

export function MembershipCard(props: MembershipCardProps) {
  const { fKey, organizationFragmentRef } = props;
  const { t } = useTranslation();

  const { membership, ...user } = useFragment<MembershipCardFragment$key>(
    fragment,
    fKey,
  );
  const organization = useFragment<MembershipCard_organizationFragment$key>(
    organizationFragment,
    organizationFragmentRef,
  );
  const isExpired
    = membership.lastSession && parseDate(membership.lastSession.expiresAt) < new Date();
  const isAssuming = !!membership.lastSession && !isExpired;

  const getAuthBadge = () => {
    if (isAssuming) {
      return (
        <Badge variant="success" className="flex items-center gap-1">
          <IconCheckmark1 size={14} />
          {t("membershipCard.status.authenticated")}
        </Badge>
      );
    } else if (isExpired) {
      return (
        <Badge variant="warning" className="flex items-center gap-1">
          <IconClock size={14} />
          {t("membershipCard.status.sessionExpired")}
        </Badge>
      );
    } else {
      return (
        <Badge variant="neutral" className="flex items-center gap-1">
          <IconLock size={14} />
          {t("membershipCard.status.authenticationRequired")}
        </Badge>
      );
    }
  };

  return (
    <Card padded className="w-full">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 hover:text-primary flex-1">
          <Avatar
            src={organization.logo?.downloadUrl}
            name={organization.name}
            size="l"
          />
          <div className="flex flex-col gap-1">
            <h2 className="font-semibold text-xl">{organization.name}</h2>
            {getAuthBadge()}
          </div>
        </div>
        <div className="flex items-center gap-3">
          {user.state === "ACTIVE"
            ? (
                <Link to={`/organizations/${organization.id}`}>
                  {isAssuming
                    ? <Button variant="secondary">{t("membershipCard.actions.start")}</Button>
                    : <Button>{t("membershipCard.actions.login")}</Button>}
                </Link>
              )
            : (
                <Button variant="secondary" disabled>{t("membershipCard.accountDeactivated")}</Button>
              )}
        </div>
      </div>
    </Card>
  );
}
