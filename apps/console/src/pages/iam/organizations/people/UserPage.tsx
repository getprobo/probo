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

import { ArchiveIcon, CaretLeftIcon, DotsThreeVerticalIcon, EnvelopeIcon, TrashIcon } from "@phosphor-icons/react";
import { getRole } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { ProfileState, UserPageQuery } from "#/__generated__/iam/UserPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { DeactivateUserDialog } from "./_components/DeactivateUserDialog";
import { RemoveUserDialog } from "./_components/RemoveUserDialog";
import { SendActivationEmailDialog } from "./_components/SendActivationEmailDialog";
import { UserPropertiesSection } from "./_components/UserPropertiesSection";
import { UserRoleSelect } from "./_components/UserRoleSelect";
import { userPage } from "./variants";

export const userPageQuery = graphql`
  query UserPageQuery($personId: ID!) {
    person: node(id: $personId) @required(action: THROW) {
      __typename
      ... on Profile {
        id
        fullName
        emailAddress
        source
        state
        kind
        createdAt
        avatar {
          downloadUrl
        }
        membership @required(action: THROW) {
          ...UserRoleSelect_membership
        }
        contract {
          start
          end
        }
        pendingInvitations(first: 1) @required(action: THROW) {
          edges {
            __typename
          }
        }
        canDeactivate: permission(action: "iam:membership-profile:deactivate")
        canRemoveMember: permission(action: "iam:membership:delete")
        canInvite: permission(action: "iam:invitation:create")
        ...SendActivationEmailDialog_profile
        ...DeactivateUserDialog_profile
        ...RemoveUserDialog_profile
        ...UserPropertiesSection_profile
      }
    }
  }
`;

function statusBadgeColor(state: ProfileState) {
  if (state === "ACTIVE") {
    return "green" as const;
  }
  if (state === "PENDING") {
    return "amber" as const;
  }
  return "neutral" as const;
}

interface UserPageProps {
  queryRef: PreloadedQuery<UserPageQuery>;
}

export function UserPage({ queryRef }: UserPageProps) {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { person } = usePreloadedQuery<UserPageQuery>(userPageQuery, queryRef);
  if (person.__typename !== "Profile") {
    throw new NotFoundError(t("userPage.notFound"));
  }

  usePageTitle(person.fullName);

  const [sendOpen, setSendOpen] = useState(false);
  const [deactivateOpen, setDeactivateOpen] = useState(false);
  const [removeOpen, setRemoveOpen] = useState(false);
  const isActive = person.state === "ACTIVE";
  const isInactive = person.state === "DEACTIVATED";
  const canSendActivationMail = !isActive && person.source !== "SCIM" && person.canInvite;
  const canDeactivate = person.canDeactivate && person.source !== "SCIM" && person.state !== "DEACTIVATED";
  const canRemove = person.canRemoveMember && person.source !== "SCIM";
  const hasActions = canSendActivationMail || canDeactivate || canRemove;
  const isResend = person.pendingInvitations.edges.length > 0;
  const showSource = person.source === "SCIM" || person.source === "SAML";
  const hasContractDates = person.contract?.start != null || person.contract?.end != null;
  const dateLabel = hasContractDates
    ? t("usersList.contract.range", {
        start: person.contract?.start
          ? dateFormat(i18n.language, person.contract.start)
          : t("usersList.contract.empty"),
        end: person.contract?.end
          ? dateFormat(i18n.language, person.contract.end)
          : t("usersList.contract.empty"),
      })
    : t("usersList.created", { date: dateFormat(i18n.language, person.createdAt) });
  const {
    root,
    back,
    header,
    person: personSlot,
    avatar,
    source,
    identity,
    kind,
    title,
    email,
    badges,
    menu,
    meta,
    contract,
  } = userPage({ inactive: isInactive });

  function handleLeft() {
    void navigate("..");
  }

  return (
    <div className={root()}>
      <Link
        to=".."
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("userPage.back")}
      </Link>
      {isInactive && (
        <Callout color="amber">
          {t("userPage.deactivatedCallout")}
        </Callout>
      )}
      <div className={header()}>
        <div className={personSlot()}>
          <div className={avatar()}>
            <Avatar
              name={person.fullName}
              email={person.emailAddress}
              src={person.avatar?.downloadUrl}
              size={5}
            />
            {showSource && (
              <span className={source()}>
                <Badge variant="soft" color="neutral" size={1}>
                  {person.source}
                </Badge>
              </span>
            )}
          </div>
          <div className={identity()}>
            {person.kind != null && (
              <Text size={1} color="faint" className={kind()}>
                {getRole(t, person.kind)}
              </Text>
            )}
            <Heading level={1} size={6} weight="medium" highContrast className={title()}>
              {person.fullName}
            </Heading>
            <Text size={2} className={email()}>
              {person.emailAddress}
            </Text>
            <div className={badges()}>
              <Badge variant="soft" color={statusBadgeColor(person.state)} size={1}>
                {t(`usersList.filters.${person.state.toLowerCase()}`)}
              </Badge>
            </div>
          </div>
        </div>
        {hasActions && (
          <div className={menu()}>
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton
                    variant="ghost"
                    color="neutral"
                    size={1}
                    aria-label={t("userListItem.actions.more")}
                  >
                    <DotsThreeVerticalIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup align="end">
                {canSendActivationMail && (
                  <DropdownItem
                    iconStart={<EnvelopeIcon />}
                    onClick={() => setSendOpen(true)}
                  >
                    {isResend
                      ? t("userListItem.actions.resendActivationMail")
                      : t("userListItem.actions.sendActivationMail")}
                  </DropdownItem>
                )}
                {!canSendActivationMail && canDeactivate && (
                  <DropdownItem
                    iconStart={<ArchiveIcon />}
                    onClick={() => setDeactivateOpen(true)}
                  >
                    {t("userListItem.actions.deactivatePerson")}
                  </DropdownItem>
                )}
                {canRemove && (
                  <DropdownItem
                    color="error"
                    iconStart={<TrashIcon />}
                    onClick={() => setRemoveOpen(true)}
                  >
                    {t("userListItem.actions.removePerson")}
                  </DropdownItem>
                )}
              </DropdownPopup>
            </Dropdown>
          </div>
        )}
      </div>
      <div className={meta()}>
        <UserRoleSelect membershipKey={person.membership} />
        <Text size={1} color="faint" className={contract()}>
          {dateLabel}
        </Text>
      </div>
      <UserPropertiesSection profileKey={person} />
      {canSendActivationMail && (
        <SendActivationEmailDialog
          profileKey={person}
          open={sendOpen}
          onOpenChange={setSendOpen}
        />
      )}
      {canDeactivate && (
        <DeactivateUserDialog
          profileKey={person}
          open={deactivateOpen}
          onOpenChange={setDeactivateOpen}
          onDeactivated={handleLeft}
        />
      )}
      {canRemove && (
        <RemoveUserDialog
          profileKey={person}
          open={removeOpen}
          onOpenChange={setRemoveOpen}
          onRemoved={handleLeft}
        />
      )}
    </div>
  );
}
