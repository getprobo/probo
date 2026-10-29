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

import { CompassIcon } from "@phosphor-icons/react";
import { useHotkeys } from "@probo/hotkeys";
import { Combobox } from "@probo/ui/src/v2/Combobox/Combobox";
import { ComboboxEmpty } from "@probo/ui/src/v2/Combobox/ComboboxEmpty";
import { ComboboxInput } from "@probo/ui/src/v2/Combobox/ComboboxInput";
import { ComboboxInputGroup } from "@probo/ui/src/v2/Combobox/ComboboxInputGroup";
import { ComboboxItem } from "@probo/ui/src/v2/Combobox/ComboboxItem";
import { ComboboxList } from "@probo/ui/src/v2/Combobox/ComboboxList";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { shortcutAria } from "@probo/ui/src/v2/shortcutAria";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useNavigate } from "react-router";

import type { navPermissions_organization$key } from "#/__generated__/iam/navPermissions_organization.graphql";
import type { NavSpotlight_organization$key } from "#/__generated__/iam/NavSpotlight_organization.graphql";
import { SearchShortcutHint } from "#/components/SearchShortcutHint/SearchShortcutHint";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { visibleNavDestinations } from "#/pages/iam/organizations/_lib/navDestinations";
import {
  navGroupByKey,
  navHref,
} from "#/pages/iam/organizations/_lib/navigation";
import { navPermissionsFragment } from "#/pages/iam/organizations/_lib/navPermissions";

import { navSpotlight } from "./variants";

const navSpotlightFragment = graphql`
  fragment NavSpotlight_organization on Organization {
    ...navPermissions_organization
  }
`;

interface SpotlightDestination {
  id: string;
  label: string;
  groupLabel: string;
  to: string;
}

export interface NavSpotlightProps {
  organizationKey: NavSpotlight_organization$key;
  slackbotAvailable: boolean;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function NavSpotlight({
  organizationKey,
  slackbotAvailable,
  open,
  onOpenChange,
}: NavSpotlightProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const organization = useFragment(navSpotlightFragment, organizationKey);
  const permissions = useFragment<navPermissions_organization$key>(
    navPermissionsFragment,
    organization,
  );
  const slots = navSpotlight();

  const destinations = useMemo(() => {
    return visibleNavDestinations(permissions, slackbotAvailable).map(
      (destination) => {
        const group = navGroupByKey(destination.group);
        return {
          id: destination.id,
          label: t(destination.labelKey),
          groupLabel: t(`nav.groups.${destination.group}`),
          to: navHref(organizationId, group, destination.path),
        };
      },
    );
  }, [organizationId, permissions, slackbotAvailable, t]);

  useHotkeys("mod+k", () => {
    onOpenChange(!open);
  }, { preventDefault: true });

  function onValueChange(destination: SpotlightDestination | null) {
    if (destination == null) {
      return;
    }
    void navigate(destination.to);
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup placement="top" lockScroll>
        <DialogTitle className={slots.title()}>{t("nav.spotlight.title")}</DialogTitle>
        <DialogBody>
          <Combobox<SpotlightDestination>
            items={destinations}
            value={null}
            onValueChange={onValueChange}
            itemToStringLabel={destination => `${destination.label} ${destination.groupLabel}`}
            inline
            open={open}
            onOpenChange={onOpenChange}
            autoHighlight
          >
            <ComboboxInputGroup className={slots.field()}>
              <span className={slots.icon()}>
                <CompassIcon />
              </span>
              <ComboboxInput
                placeholder={t("nav.spotlight.placeholder")}
                aria-keyshortcuts={shortcutAria("K")}
              />
              <span className={slots.hint()}>
                <SearchShortcutHint shortcut="K" />
              </span>
            </ComboboxInputGroup>
            <ComboboxEmpty>{t("nav.spotlight.empty")}</ComboboxEmpty>
            <ComboboxList<SpotlightDestination> className={slots.list()}>
              {destination => (
                <ComboboxItem key={destination.id} value={destination}>
                  <span className={slots.row()}>
                    <span>{destination.label}</span>
                    <span className={slots.group()}>{destination.groupLabel}</span>
                  </span>
                </ComboboxItem>
              )}
            </ComboboxList>
          </Combobox>
        </DialogBody>
      </DialogPopup>
    </Dialog>
  );
}
