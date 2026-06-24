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
