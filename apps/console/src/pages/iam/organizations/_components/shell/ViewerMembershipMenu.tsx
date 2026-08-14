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

import {
  CaretDownIcon,
  CaretRightIcon,
  FileTextIcon,
  KeyIcon,
  MoonIcon,
  QuestionIcon,
  SignOutIcon,
  SunIcon,
  UserIcon,
} from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import { useDisplayMode } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownGroup } from "@probo/ui/src/v2/Dropdown/DropdownGroup";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";
import { Link } from "react-router";

import type { ViewerMembershipMenu_organization$key } from "#/__generated__/iam/ViewerMembershipMenu_organization.graphql";
import type { ViewerMembershipMenuSignOutMutation } from "#/__generated__/iam/ViewerMembershipMenuSignOutMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { navRail, viewerMembershipMenuTrigger } from "./variants";

const viewerMembershipMenuFragment = graphql`
  fragment ViewerMembershipMenu_organization on Organization {
    viewer @required(action: THROW) {
      fullName
      identity @required(action: THROW) {
        email
        canListAPIKeys: permission(action: "iam:personal-api-key:list")
        canListOAuth2AccessTokens: permission(
          action: "iam:oauth2-access-token:list"
        )
      }
    }
  }
`;

const signOutMutation = graphql`
  mutation ViewerMembershipMenuSignOutMutation {
    signOut {
      success
    }
  }
`;

export interface ViewerMembershipMenuProps {
  // "rail" is a full-width row matching the nav items, avatar first and the
  // name revealed as the rail opens; "bar" is the pill the employee portal's
  // top bar uses.
  variant?: "bar" | "rail";
  organizationKey: ViewerMembershipMenu_organization$key;
}

/** Who you are signed in as, and the account-level actions. */
export function ViewerMembershipMenu({ variant = "bar", organizationKey }: ViewerMembershipMenuProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { toast } = useToast();
  const { displayMode, toggleDisplayMode } = useDisplayMode();

  const {
    viewer: {
      fullName,
      identity: { canListAPIKeys, canListOAuth2AccessTokens, email },
    },
  } = useFragment(viewerMembershipMenuFragment, organizationKey);
  const [signOut, isSigningOut] = useMutation<ViewerMembershipMenuSignOutMutation>(signOutMutation);

  // Someone invited but not yet onboarded may have no name yet.
  const displayName = fullName.trim() || email;

  const handleSignOut = () => {
    signOut({
      variables: {},
      onCompleted: (_, errors) => {
        if (errors) {
          toast({
            title: t("viewerMembershipDropdown.errors.requestFailed"),
            description: formatError(t("viewerMembershipDropdown.errors.cannotSignOut"), errors),
            variant: "error",
          });
          return;
        }
        // Full reload rather than a client navigation, so no Relay store
        // survives the session it belonged to.
        window.location.href = "/auth/login";
      },
      onError: (error) => {
        toast({
          title: t("common.error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  const railSlots = navRail();
  const isRail = variant === "rail";

  const trigger = isRail
    ? (
        <button type="button" className={railSlots.item()}>
          <span className={railSlots.icon()}>
            {/* A person reads as a round avatar, against the organization's
                squared one at the top of the rail. */}
            <Avatar
              size={2}
              variant="soft"
              color="gold"
              radius="full"
              fallback={displayName.charAt(0).toUpperCase() || <UserIcon />}
            />
          </span>
          <Text size={2} weight="medium" color="neutral" highContrast className={railSlots.label()}>
            {displayName}
          </Text>
          {/* Points right, matching where the menu opens. */}
          <CaretRightIcon className={railSlots.caret()} />
        </button>
      )
    : (
        <button type="button" className={viewerMembershipMenuTrigger()} aria-label={displayName}>
          <Avatar size={1} variant="soft" color="gold" radius="small" fallback={<UserIcon />} />
          <Text size={2} weight="medium" color="neutral" highContrast className="max-w-36 truncate">
            {displayName}
          </Text>
          <CaretDownIcon className="size-4 shrink-0 text-sand-11" />
        </button>
      );

  return (
    <Dropdown>
      <DropdownTrigger render={trigger} />
      {/* Anchored to the side in the rail so the menu clears it rather than
          covering the column it was opened from. The offset is measured from
          the trigger, which sits inside the rail's 8px padding, so it takes
          that plus a gap to clear the rail's edge. */}
      <DropdownPopup
        side={isRail ? "right" : "bottom"}
        sideOffset={isRail ? 12 : 4}
        align="end"
      >
        <DropdownGroup>
          <div className="flex w-full flex-col gap-1 px-3 py-3">
            <Text size={2} weight="medium" color="neutral" highContrast>
              {displayName}
            </Text>
            <Text size={1} color="faint" className="truncate">
              {email}
            </Text>
          </div>
        </DropdownGroup>
        <DropdownSeparator />
        {canListAPIKeys && (
          <DropdownItem iconStart={<KeyIcon />} render={<Link to="/me/api-keys" />}>
            {t("viewerMembershipDropdown.actions.apiKeys")}
          </DropdownItem>
        )}
        {canListOAuth2AccessTokens && (
          <DropdownItem iconStart={<KeyIcon />} render={<Link to="/me/oauth-tokens" />}>
            {t("viewerMembershipDropdown.actions.oauthTokens")}
          </DropdownItem>
        )}
        <DropdownItem
          iconStart={<FileTextIcon />}
          render={<Link to={`/organizations/${organizationId}/employee`} />}
        >
          {t("viewerMembershipDropdown.actions.employeePortal")}
        </DropdownItem>
        {!isRail && (
          <>
            <DropdownItem
              iconStart={displayMode === "dark" ? <SunIcon /> : <MoonIcon />}
              onClick={toggleDisplayMode}
            >
              {displayMode === "dark"
                ? t("viewerMembershipDropdown.actions.switchToLightMode")
                : t("viewerMembershipDropdown.actions.switchToDarkMode")}
            </DropdownItem>
            <DropdownItem iconStart={<QuestionIcon />} render={<a href="mailto:support@probo.com" />}>
              {t("viewerMembershipDropdown.actions.help")}
            </DropdownItem>
          </>
        )}
        <DropdownSeparator />
        <DropdownItem
          color="error"
          iconStart={<SignOutIcon />}
          disabled={isSigningOut}
          onClick={handleSignOut}
        >
          {t("viewerMembershipDropdown.actions.logout")}
        </DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}
