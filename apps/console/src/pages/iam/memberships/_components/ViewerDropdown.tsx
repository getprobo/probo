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
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ViewerDropdownFragment$key } from "#/__generated__/iam/ViewerDropdownFragment.graphql";
import type { ViewerDropdownSignOutMutation } from "#/__generated__/iam/ViewerDropdownSignOutMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { topBarUserMenuTrigger } from "./TopBar/variants";

export const fragment = graphql`
  fragment ViewerDropdownFragment on Identity {
    canListOAuth2AccessTokens: permission(action: "iam:oauth2-access-token:list")
    email
    fullName
  }
`;

const signOutMutation = graphql`
  mutation ViewerDropdownSignOutMutation {
    signOut {
      success
    }
  }
`;

interface ViewerDropdownProps {
  identityKey: ViewerDropdownFragment$key;
}

export function ViewerDropdown({ identityKey }: ViewerDropdownProps) {
  const { t } = useTranslation();
  const { displayMode, toggleDisplayMode } = useDisplayMode();

  const { canListOAuth2AccessTokens, email, fullName }
    = useFragment<ViewerDropdownFragment$key>(fragment, identityKey);
  const [signOut, isSigningOut] = useMutation<ViewerDropdownSignOutMutation>(
    signOutMutation,
    { errorToast: t("viewerDropdown.errors.cannotSignOut") },
  );

  const displayName = fullName.trim() || email;

  function handleSignOut() {
    void signOut({ variables: {} }).then(() => {
      // Full reload rather than a client navigation, so no Relay store
      // survives the session it belonged to.
      window.location.href = "/auth/login";
    }).catch(() => {
      // errorToast already handles user-facing feedback.
    });
  }

  return (
    <Dropdown>
      <DropdownTrigger
        render={(
          <button type="button" className={topBarUserMenuTrigger()} aria-label={displayName}>
            <Avatar
              size={1}
              variant="soft"
              color="gold"
              radius="small"
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
        {canListOAuth2AccessTokens && (
          <DropdownItem iconStart={<KeyIcon />} render={<Link to="/me/oauth-tokens" />}>
            {t("viewerDropdown.actions.oauthTokens")}
          </DropdownItem>
        )}
        <DropdownItem
          iconStart={displayMode === "dark" ? <SunIcon /> : <MoonIcon />}
          onClick={toggleDisplayMode}
        >
          {displayMode === "dark"
            ? t("nav.switchToLightMode")
            : t("nav.switchToDarkMode")}
        </DropdownItem>
        <DropdownItem iconStart={<QuestionIcon />} render={<a href="mailto:support@probo.com" />}>
          {t("viewerDropdown.actions.help")}
        </DropdownItem>
        <DropdownSeparator />
        <DropdownItem
          color="error"
          iconStart={<SignOutIcon />}
          disabled={isSigningOut}
          onClick={handleSignOut}
        >
          {t("viewerDropdown.actions.logout")}
        </DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}
