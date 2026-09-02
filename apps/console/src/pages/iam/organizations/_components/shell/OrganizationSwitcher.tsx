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

import { CaretDownIcon, MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useQueryLoader } from "react-relay";
import { Link, useLocation } from "react-router";

import type { OrganizationSwitcher_organization$key } from "#/__generated__/iam/OrganizationSwitcher_organization.graphql";
import type { OrganizationSwitcherMenuQuery } from "#/__generated__/iam/OrganizationSwitcherMenuQuery.graphql";

import {
  OrganizationSwitcherMenu,
  organizationSwitcherMenuQuery,
} from "./OrganizationSwitcherMenu";
import { navRail, organizationSwitcher } from "./variants";

const organizationSwitcherFragment = graphql`
  fragment OrganizationSwitcher_organization on Organization {
    name
    logo {
      downloadUrl
    }
  }
`;

export interface OrganizationSwitcherProps {
  organizationKey: OrganizationSwitcher_organization$key;
}

export function OrganizationSwitcher({ organizationKey }: OrganizationSwitcherProps) {
  const { t } = useTranslation();
  const location = useLocation();
  const [search, setSearch] = useState("");

  const organization = useFragment(organizationSwitcherFragment, organizationKey);
  const [queryRef, loadQuery] = useQueryLoader<OrganizationSwitcherMenuQuery>(
    organizationSwitcherMenuQuery,
  );

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadQuery({}, { fetchPolicy: "network-only" });
      return;
    }
    setSearch("");
  }, [loadQuery]);

  const slots = organizationSwitcher();
  const railSlots = navRail();

  const trigger = (
    <button type="button" className={railSlots.item()}>
      <span className={railSlots.icon()}>
        <Avatar
          size={2}
          variant="solid"
          color="gold"
          radius="small"
          src={organization.logo?.downloadUrl ?? undefined}
          fallback={organization.name.charAt(0) || "?"}
        />
      </span>
      <Text size={2} weight="medium" color="neutral" highContrast className={railSlots.label()}>
        {organization.name}
      </Text>
      <CaretDownIcon className={railSlots.caret()} />
    </button>
  );

  return (
    <Dropdown onOpenChange={handleOpenChange}>
      <DropdownTrigger render={trigger} />
      <DropdownPopup side="bottom" align="start" className={slots.popup()}>
        <div className={slots.search()}>
          <TextField
            autoFocus
            size={1}
            icon={<MagnifyingGlassIcon />}
            placeholder={t("membershipsDropdown.searchPlaceholder")}
            value={search}
            onValueChange={setSearch}
            // The menu treats typing as type-ahead navigation, which would
            // otherwise swallow most of what is typed into the search box.
            onKeyDown={(event) => {
              if (event.key.length === 1 || event.key === "Backspace" || event.key === "Delete") {
                event.stopPropagation();
              }
            }}
          />
        </div>
        <div className={slots.list()}>
          {queryRef != null && (
            <Suspense
              fallback={(
                <Text size={2} color="faint" className="block px-3 py-2">
                  {t("membershipsDropdown.loading")}
                </Text>
              )}
            >
              <OrganizationSwitcherMenu search={search} queryRef={queryRef} />
            </Suspense>
          )}
        </div>
        <DropdownSeparator />
        <DropdownItem
          iconStart={<PlusIcon />}
          render={(
            <Link
              to="/organizations/new"
              state={{ from: `${location.pathname}${location.search}${location.hash}` }}
            />
          )}
        >
          {t("membershipsDropdown.addOrganization")}
        </DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}
