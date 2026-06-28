// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { CaretDownIcon, SignOutIcon, UserIcon } from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownGroup } from "@probo/ui/src/v2/Dropdown/DropdownGroup";
import { DropdownGroupLabel } from "@probo/ui/src/v2/Dropdown/DropdownGroupLabel";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TopBarUserMenu_identity$key } from "./__generated__/TopBarUserMenu_identity.graphql";
import { topBarUserMenuTrigger } from "./variants";

const topBarUserMenuFragment = graphql`
  fragment TopBarUserMenu_identity on Identity {
    fullName
    email
  }
`;

interface TopBarUserMenuProps {
  identityKey: TopBarUserMenu_identity$key;
}

export function TopBarUserMenu({ identityKey }: TopBarUserMenuProps) {
  const { t } = useTranslation();
  const identity = useFragment(topBarUserMenuFragment, identityKey);

  return (
    <Dropdown>
      <DropdownTrigger
        render={(
          <button type="button" className={topBarUserMenuTrigger()} aria-label={identity.fullName}>
            <Avatar
              size={1}
              variant="soft"
              color="gold"
              radius="small"
              fallback={<UserIcon />}
            />
            <Text size={2} weight="medium" color="neutral" highContrast>
              {identity.fullName}
            </Text>
            <CaretDownIcon className="size-4 text-sand-11" />
          </button>
        )}
      />
      <DropdownPopup align="end">
        <DropdownGroup>
          <DropdownGroupLabel>{identity.email}</DropdownGroupLabel>
        </DropdownGroup>
        <DropdownSeparator />
        <DropdownItem color="error" iconStart={<SignOutIcon />}>
          {t("userMenu.signOut")}
        </DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}
