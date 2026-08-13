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

import { CaretDownIcon, CaretUpDownIcon, MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
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
  // "rail" is a full-width row matching the nav items, avatar first and the
  // name revealed as the rail opens; "bar" is the pill the employee portal's
  // top bar uses.
  variant?: "bar" | "rail";
  organizationKey: OrganizationSwitcher_organization$key;
}

/**
 * Names the current organization and switches to another.
 *
 * The list is only fetched when the menu opens: most sessions never switch, and
 * the query walks every organization the viewer belongs to.
 */
export function OrganizationSwitcher({ variant = "bar", organizationKey }: OrganizationSwitcherProps) {
  const { t } = useTranslation();
  const location = useLocation();
  const [search, setSearch] = useState("");

  const organization = useFragment(organizationSwitcherFragment, organizationKey);
  const [queryRef, loadQuery] = useQueryLoader<OrganizationSwitcherMenuQuery>(
    organizationSwitcherMenuQuery,
  );

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadQuery({});
    }
  }, [loadQuery]);

  const slots = organizationSwitcher();
  const railSlots = navRail();
  const isRail = variant === "rail";

  const trigger = isRail
    ? (
        <button type="button" className={railSlots.item()}>
          <span className={railSlots.icon()}>
            {/* One step below the 40px icon box, so the row's hover reads as a
                ring around the mark rather than stopping at its edge. Solid
                gold so an organization with no logo still shows a mark rather
                than faint text. */}
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
          {/* Points down, matching where the menu opens. */}
          <CaretDownIcon className={railSlots.caret()} />
        </button>
      )
    : (
        <button type="button" className={slots.trigger()}>
          <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
            {organization.name}
          </Text>
          <CaretUpDownIcon className="size-4 shrink-0 text-sand-11" />
        </button>
      );

  return (
    <Dropdown onOpenChange={handleOpenChange}>
      <DropdownTrigger render={trigger} />
      {/* Drops below the trigger in both presentations. In the rail that puts
          it just inside the left edge, offset by the rail's own padding. */}
      <DropdownPopup side="bottom" align="start" className={slots.popup()}>
        <div className={slots.search()}>
          <TextField
            autoFocus
            size={1}
            icon={<MagnifyingGlassIcon />}
            placeholder={t("membershipsDropdown.searchPlaceholder")}
            value={search}
            onValueChange={setSearch}
            // The menu treats typing as type-ahead navigation; without this the
            // search box would lose most of what is typed into it.
            onKeyDown={(event) => { event.stopPropagation(); }}
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
          render={<Link to="/organizations/new" state={{ from: location.pathname }} />}
        >
          {t("membershipsDropdown.addOrganization")}
        </DropdownItem>
      </DropdownPopup>
    </Dropdown>
  );
}
