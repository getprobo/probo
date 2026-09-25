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

import { ArchiveIcon, DotsThreeVerticalIcon, EnvelopeIcon, TrashIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { UserListItem_profile$key } from "#/__generated__/iam/UserListItem_profile.graphql";

import { profileStateBadgeColor } from "../_lib/profileStateBadgeColor";
import { isUsersListKind } from "../_lib/useUsersListFilters";
import { userListItem } from "../variants";

import { DeactivateUserDialog } from "./DeactivateUserDialog";
import { RemoveUserDialog } from "./RemoveUserDialog";
import { SendActivationEmailDialog } from "./SendActivationEmailDialog";
import { UserRoleSelect } from "./UserRoleSelect";

const fragment = graphql`
  fragment UserListItem_profile on Profile {
    id
    source
    state
    kind
    fullName
    emailAddress
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
    lastInvitation: pendingInvitations(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
    @required(action: THROW)
    @connection(key: "SendActivationEmailDialog_lastInvitation") {
      edges {
        __typename
      }
    }
    canInvite: permission(action: "iam:invitation:create")
    canDeactivate: permission(action: "iam:membership-profile:deactivate")
    canRemoveMember: permission(action: "iam:membership:delete")
    ...SendActivationEmailDialog_profile
    ...DeactivateUserDialog_profile
    ...RemoveUserDialog_profile
  }
`;

interface UserListItemProps {
  profileKey: UserListItem_profile$key;
  onDeactivated: () => void;
  onDeleted: () => void;
}

export function UserListItem({ profileKey, onDeactivated, onDeleted }: UserListItemProps) {
  const { t, i18n } = useTranslation();
  const profile = useFragment(fragment, profileKey);
  const [sendOpen, setSendOpen] = useState(false);
  const [deactivateOpen, setDeactivateOpen] = useState(false);
  const [removeOpen, setRemoveOpen] = useState(false);
  const isActive = profile.state === "ACTIVE";
  const isInactive = profile.state === "DEACTIVATED";
  const canSendActivationMail = !isActive && profile.source !== "SCIM" && profile.canInvite;
  const canDeactivate = profile.canDeactivate && profile.source !== "SCIM" && profile.state !== "DEACTIVATED";
  const canRemove = profile.canRemoveMember && profile.source !== "SCIM";
  const hasActions = canSendActivationMail || canDeactivate || canRemove;
  const isResend = profile.lastInvitation.edges.length > 0;
  const {
    card,
    overlay,
    person,
    avatar,
    source,
    menu,
    identity,
    kind,
    title,
    email,
    meta,
    metaRow,
    contract,
  } = userListItem({ inactive: isInactive });
  const showSource = profile.source === "SCIM" || profile.source === "SAML";
  const hasContractDates = profile.contract?.start != null || profile.contract?.end != null;
  const dateLabel = hasContractDates
    ? t("usersList.contract.range", {
        start: profile.contract?.start
          ? dateFormat(i18n.language, profile.contract.start)
          : t("usersList.contract.empty"),
        end: profile.contract?.end
          ? dateFormat(i18n.language, profile.contract.end)
          : t("usersList.contract.empty"),
      })
    : t("usersList.created", { date: dateFormat(i18n.language, profile.createdAt) });

  return (
    <Card variant="soft" size={2} padding="none" interactive className={card()}>
      <Link to={profile.id} underline={false} className={overlay()} aria-label={profile.fullName} />
      <div className={person()}>
        <div className={avatar()}>
          <Avatar
            name={profile.fullName}
            email={profile.emailAddress}
            src={profile.avatar?.downloadUrl}
            size={4}
          />
          {showSource && (
            <span className={source()}>
              <Badge variant="soft" color="neutral" size={1}>
                {profile.source}
              </Badge>
            </span>
          )}
        </div>
        <div className={identity()}>
          {profile.kind != null && (
            <Text size={1} color="faint" className={kind()}>
              {isUsersListKind(profile.kind) ? t(`userForm.kinds.${profile.kind}`) : profile.kind}
            </Text>
          )}
          <Heading level={2} size={3} weight="medium" highContrast className={title()}>
            {profile.fullName}
          </Heading>
          <Text size={2} className={email()}>
            {profile.emailAddress}
          </Text>
        </div>
      </div>
      <Separator />
      <div className={meta()}>
        <div className={metaRow()}>
          <Badge variant="soft" color={profileStateBadgeColor(profile.state)} size={1}>
            {t(`usersList.filters.${profile.state.toLowerCase()}`)}
          </Badge>
          <UserRoleSelect membershipKey={profile.membership} />
        </div>
        <Text size={1} color="faint" className={contract()}>
          {dateLabel}
        </Text>
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
              {canDeactivate && (
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
      {canSendActivationMail && (
        <SendActivationEmailDialog
          profileKey={profile}
          open={sendOpen}
          onOpenChange={setSendOpen}
        />
      )}
      {canDeactivate && (
        <DeactivateUserDialog
          profileKey={profile}
          open={deactivateOpen}
          onOpenChange={setDeactivateOpen}
          onDeactivated={onDeactivated}
        />
      )}
      {canRemove && (
        <RemoveUserDialog
          profileKey={profile}
          open={removeOpen}
          onOpenChange={setRemoveOpen}
          onRemoved={onDeleted}
        />
      )}
    </Card>
  );
}
