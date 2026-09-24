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
import { usePageTitle } from "@probo/hooks";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { ProfileState, UserPageQuery } from "#/__generated__/iam/UserPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { DeactivateUserDialog } from "./_components/DeactivateUserDialog";
import { RemoveUserDialog } from "./_components/RemoveUserDialog";
import { SendActivationEmailDialog } from "./_components/SendActivationEmailDialog";
import { UserIdentitySection } from "./_components/UserIdentitySection";
import { UserPropertiesSection } from "./_components/UserPropertiesSection";
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
        ...UserIdentitySection_profile
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

interface ExtraAction {
  key: string;
  label: string;
  icon: ReactNode;
  color?: "error";
  onSelect: () => void;
}

interface UserPageProps {
  queryRef: PreloadedQuery<UserPageQuery>;
}

export function UserPage({ queryRef }: UserPageProps) {
  const { t } = useTranslation();
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
  const isResend = person.pendingInvitations.edges.length > 0;
  const { root, back, header, identity, titleRow, title, email, actions } = userPage();
  const extraActions: ExtraAction[] = [];
  if (canRemove) {
    extraActions.push({
      key: "remove",
      label: t("userListItem.actions.removePerson"),
      icon: <TrashIcon />,
      color: "error",
      onSelect: () => setRemoveOpen(true),
    });
  }
  const onlyExtra = extraActions.length === 1 ? extraActions[0] : null;
  const hasToolbarActions = canSendActivationMail || canDeactivate || extraActions.length > 0;

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
      <div className={header()}>
        <div className={identity()}>
          <div className={titleRow()}>
            <Heading level={1} size={6} weight="medium" highContrast className={title()}>
              {person.fullName}
            </Heading>
            <Badge variant="soft" color={statusBadgeColor(person.state)} size={1}>
              {t(`usersList.filters.${person.state.toLowerCase()}`)}
            </Badge>
          </div>
          <Text size={2} className={email()}>
            {person.emailAddress}
          </Text>
        </div>
        {hasToolbarActions && (
          <div className={actions()}>
            {canSendActivationMail && (
              <Button
                type="button"
                size={2}
                variant="solid"
                color="neutral"
                highContrast
                iconStart={<EnvelopeIcon />}
                onClick={() => setSendOpen(true)}
              >
                {isResend
                  ? t("userListItem.actions.resendActivationMail")
                  : t("userListItem.actions.sendActivationMail")}
              </Button>
            )}
            {!canSendActivationMail && canDeactivate && (
              <Button
                type="button"
                size={2}
                variant="solid"
                color="neutral"
                highContrast
                iconStart={<ArchiveIcon />}
                onClick={() => setDeactivateOpen(true)}
              >
                {t("userListItem.actions.deactivatePerson")}
              </Button>
            )}
            {onlyExtra != null && (
              <IconButton
                type="button"
                size={2}
                variant="surface"
                color={onlyExtra.color === "error" ? "red" : "neutral"}
                aria-label={onlyExtra.label}
                onClick={onlyExtra.onSelect}
              >
                {onlyExtra.icon}
              </IconButton>
            )}
            {extraActions.length > 1 && (
              <Dropdown>
                <DropdownTrigger
                  render={(
                    <IconButton
                      variant="surface"
                      color="neutral"
                      size={2}
                      aria-label={t("userListItem.actions.more")}
                    >
                      <DotsThreeVerticalIcon />
                    </IconButton>
                  )}
                />
                <DropdownPopup align="end">
                  {extraActions.map(action => (
                    <DropdownItem
                      key={action.key}
                      color={action.color}
                      iconStart={action.icon}
                      onClick={action.onSelect}
                    >
                      {action.label}
                    </DropdownItem>
                  ))}
                </DropdownPopup>
              </Dropdown>
            )}
          </div>
        )}
      </div>
      {isInactive && (
        <Callout color="amber">
          {t("userPage.deactivatedCallout")}
        </Callout>
      )}
      <UserIdentitySection profileKey={person} />
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
