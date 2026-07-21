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
  BellIcon,
  BellRingingIcon,
  CaretDownIcon,
  GlobeIcon,
  SignOutIcon,
  UserIcon,
} from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownGroup } from "@probo/ui/src/v2/Dropdown/DropdownGroup";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownSubmenu } from "@probo/ui/src/v2/Dropdown/DropdownSubmenu";
import { DropdownSubmenuTrigger } from "@probo/ui/src/v2/Dropdown/DropdownSubmenuTrigger";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { useSignOut } from "#/lib/auth/useSignOut";
import {
  URL_LOCALE_LABELS,
  URL_LOCALES,
  type UrlLocale,
} from "#/lib/i18n/locale";
import { useChangeLocale } from "#/lib/i18n/useChangeLocale";
import { useLocale } from "#/lib/i18n/useLocale";
import { useSubscribeDialog } from "#/lib/mailingList/subscribeDialogContext";

import type { TopBarUserMenu_identity$key } from "./__generated__/TopBarUserMenu_identity.graphql";
import { topBarUserMenuTrigger } from "./variants";

const topBarUserMenuFragment = graphql`
  fragment TopBarUserMenu_identity on Identity {
    fullName
    email
    locale
  }
`;

interface TopBarUserMenuProps {
  identityKey: TopBarUserMenu_identity$key;
}

export function TopBarUserMenu({ identityKey }: TopBarUserMenuProps) {
  const { t } = useTranslation();
  const identity = useFragment(topBarUserMenuFragment, identityKey);
  const { openSubscribe, isSubscribed, unsubscribe, isUnsubscribing } = useSubscribeDialog();
  const [signOut, isSigningOut] = useSignOut();
  const locale = useLocale();
  const [changeLocale, isChangingLocale] = useChangeLocale();

  // New users may not have set a full name yet; fall back to the email.
  const displayName = identity.fullName.trim() || identity.email;

  const onSubscribeItemClick = () => {
    if (isSubscribed) {
      void unsubscribe();
      return;
    }
    openSubscribe();
  };

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
              {identity.email}
            </Text>
          </div>
        </DropdownGroup>
        <DropdownSeparator />
        <DropdownItem
          iconStart={isSubscribed ? <BellRingingIcon /> : <BellIcon />}
          color={isSubscribed ? "success" : "accent"}
          disabled={isUnsubscribing}
          onClick={onSubscribeItemClick}
        >
          {isSubscribed ? t("userMenu.subscribed") : t("userMenu.subscribe")}
        </DropdownItem>
        <DropdownSubmenu>
          <DropdownSubmenuTrigger iconStart={<GlobeIcon />}>
            {t("locale.label")}
          </DropdownSubmenuTrigger>
          <DropdownPopup side="left" align="start">
            <DropdownRadioGroup
              value={locale}
              onValueChange={(value: string) => {
                void changeLocale(value as UrlLocale, { persist: true });
              }}
            >
              {URL_LOCALES.map(code => (
                <DropdownRadioItem key={code} value={code} disabled={isChangingLocale}>
                  {URL_LOCALE_LABELS[code]}
                </DropdownRadioItem>
              ))}
            </DropdownRadioGroup>
          </DropdownPopup>
        </DropdownSubmenu>
        <DropdownSeparator />
        <DropdownItem
          color="error"
          iconStart={<SignOutIcon />}
          disabled={isSigningOut}
          onClick={() => { void signOut(); }}
        >
          {t("userMenu.signOut")}
        </DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}
