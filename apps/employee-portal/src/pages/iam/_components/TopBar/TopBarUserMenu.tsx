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
  KeyIcon,
  MoonIcon,
  QuestionIcon,
  SignOutIcon,
  SunIcon,
  UserIcon,
} from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { useDisplayMode } from "@probo/ui/src/v2/displayMode/useDisplayMode";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownGroup } from "@probo/ui/src/v2/Dropdown/DropdownGroup";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { EditableAvatarButton } from "@probo/ui/src/v2/EditableAvatarButton/EditableAvatarButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TopBarUserMenu_identity$key } from "#/__generated__/iam/TopBarUserMenu_identity.graphql";
import type { TopBarUserMenuSignOutMutation } from "#/__generated__/iam/TopBarUserMenuSignOutMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { IdentityAvatarDialog } from "#/pages/iam/_components/IdentityAvatarDialog";

import { topBarUserMenuTrigger } from "./variants";

const topBarUserMenuFragment = graphql`
  fragment TopBarUserMenu_identity on Identity {
    email
    fullName
    avatar {
      downloadUrl
    }
    canListOAuth2AccessTokens: permission(
      action: "iam:oauth2-access-token:list"
    )
    ...IdentityAvatarDialog_identity
  }
`;

const signOutMutation = graphql`
  mutation TopBarUserMenuSignOutMutation {
    signOut {
      success
    }
  }
`;

interface TopBarUserMenuProps {
  identityKey: TopBarUserMenu_identity$key;
}

export function TopBarUserMenu({ identityKey }: TopBarUserMenuProps) {
  const { t } = useTranslation();
  const { displayMode, toggleDisplayMode } = useDisplayMode();

  const identity = useFragment(topBarUserMenuFragment, identityKey);
  const { canListOAuth2AccessTokens, email, fullName, avatar } = identity;
  const [signOut, isSigningOut] = useMutation<TopBarUserMenuSignOutMutation>(
    signOutMutation,
    { errorToast: t("userMenu.signOutFailed") },
  );
  const [avatarOpen, setAvatarOpen] = useState(false);

  const displayName = fullName.trim() || email;
  const avatarSrc = avatar?.downloadUrl;

  function handleSignOut() {
    void signOut({ variables: {} }).then(() => {
      // Full reload rather than a client navigation, so no Relay store
      // survives the session it belonged to. Site-root path — basename
      // would prefix /employee-portal.
      window.location.href = "/auth/login";
    }).catch(() => {
      // errorToast already handles user-facing feedback.
    });
  }

  return (
    <>
      <Dropdown>
        <DropdownTrigger
          render={(
            <button type="button" className={topBarUserMenuTrigger()} aria-label={displayName}>
              <Avatar
                size={1}
                variant="soft"
                color="gold"
                radius="small"
                src={avatarSrc}
                fallback={<UserIcon />}
              />
              <Text size={2} weight="medium" color="neutral" highContrast className="max-w-36 truncate">
                {displayName}
              </Text>
              <CaretDownIcon className="size-4 shrink-0 text-sand-11" />
            </button>
          )}
        />
        <DropdownPopup align="end">
          <DropdownGroup>
            <div className="flex w-full items-center gap-3 px-3 py-3">
              <EditableAvatarButton
                fullName={displayName}
                src={avatarSrc}
                fallback={<UserIcon />}
                onClick={() => setAvatarOpen(true)}
                label={t("editAvatar.actions.change")}
                size={2}
                radius="full"
              />
              <div className="flex min-w-0 flex-col gap-1">
                <Text size={2} weight="medium" color="neutral" highContrast>
                  {displayName}
                </Text>
                <Text size={1} color="faint" className="truncate">
                  {email}
                </Text>
              </div>
            </div>
          </DropdownGroup>
          <DropdownSeparator />
          {canListOAuth2AccessTokens && (
            <DropdownItem iconStart={<KeyIcon />} render={<a href="/me/oauth-tokens" />}>
              {t("userMenu.oauthTokens")}
            </DropdownItem>
          )}
          <DropdownItem
            iconStart={displayMode === "dark" ? <SunIcon /> : <MoonIcon />}
            onClick={toggleDisplayMode}
          >
            {displayMode === "dark"
              ? t("userMenu.switchToLightMode")
              : t("userMenu.switchToDarkMode")}
          </DropdownItem>
          <DropdownItem iconStart={<QuestionIcon />} render={<a href="mailto:support@probo.com" />}>
            {t("userMenu.help")}
          </DropdownItem>
          <DropdownSeparator />
          <DropdownItem
            color="error"
            iconStart={<SignOutIcon />}
            disabled={isSigningOut}
            onClick={handleSignOut}
          >
            {t("userMenu.signOut")}
          </DropdownItem>
        </DropdownPopup>
      </Dropdown>
      <IdentityAvatarDialog
        identityKey={identity}
        open={avatarOpen}
        onOpenChange={setAvatarOpen}
      />
    </>
  );
}
