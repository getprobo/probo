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
import {
  useEffect,
  useMemo,
  useState,
  useTransition,
} from "react";
import { useTranslation } from "react-i18next";
import {
  fetchQuery,
  graphql,
  useFragment,
} from "react-relay";
import { useNavigate } from "react-router";

import type { navSpotlightSearchQuery as NavSpotlightSearchQuery } from "#/__generated__/core/navSpotlightSearchQuery.graphql";
import type { navPermissions_organization$key } from "#/__generated__/iam/navPermissions_organization.graphql";
import type { NavSpotlight_organization$key } from "#/__generated__/iam/NavSpotlight_organization.graphql";
import { SearchShortcutHint } from "#/components/SearchShortcutHint/SearchShortcutHint";
import { navSpotlightSearchQuery } from "#/components/spotlight/navSpotlightSearchQuery";
import { coreEnvironment } from "#/environments";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { visibleNavDestinations } from "#/pages/iam/organizations/_lib/navDestinations";
import {
  navGroupByKey,
  navHref,
} from "#/pages/iam/organizations/_lib/navigation";
import { navPermissionsFragment } from "#/pages/iam/organizations/_lib/navPermissions";
import { resourceHref } from "#/pages/iam/organizations/_lib/resourceHref";

import { navSpotlight } from "./variants";

const navSpotlightFragment = graphql`
  fragment NavSpotlight_organization on Organization {
    ...navPermissions_organization
  }
`;

type SpotlightItem = {
  id: string;
  kind: "page" | "resource";
  label: string;
  groupLabel: string;
  to: string;
};

export interface NavSpotlightProps {
  organizationKey: NavSpotlight_organization$key;
  slackbotAvailable: boolean;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function hasQualifier(input: string): boolean {
  return /(?:^|\s)(?:tag|type):/.test(input);
}

type SearchHitNode
  = NavSpotlightSearchQuery["response"]["search"]["edges"][number]["node"];

function searchHitParentId(node: SearchHitNode): string | undefined {
  switch (node.__typename) {
    case "Control":
      return node.framework?.id;
    case "DataProtectionImpactAssessment":
    case "TransferImpactAssessment":
      return node.processingActivity?.id;
    case "TreatmentPlan":
      return node.riskAnalysis?.id;
    default:
      return undefined;
  }
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
  const [input, setInput] = useState("");
  const [resources, setResources] = useState<SpotlightItem[]>([]);
  const [, startTransition] = useTransition();

  const destinations = useMemo(() => {
    return visibleNavDestinations(permissions, slackbotAvailable).map(
      (destination) => {
        const group = navGroupByKey(destination.group);
        return {
          id: `page:${destination.id}`,
          kind: "page" as const,
          label: t(destination.labelKey),
          groupLabel: t(`nav.groups.${destination.group}`),
          to: navHref(organizationId, group, destination.path),
        };
      },
    );
  }, [
    organizationId,
    permissions,
    slackbotAvailable,
    t,
  ]);

  function handleOpenChange(next: boolean) {
    if (!next) {
      setInput("");
      setResources([]);
    }
    onOpenChange(next);
  }

  useEffect(() => {
    const trimmed = input.trim();
    if (!open || trimmed === "") {
      return;
    }

    let active = true;
    let subscription: { unsubscribe: () => void } | undefined;
    const handle = window.setTimeout(() => {
      startTransition(() => {
        subscription = fetchQuery<NavSpotlightSearchQuery>(
          coreEnvironment,
          navSpotlightSearchQuery,
          {
            organizationId,
            query: trimmed,
          },
        ).subscribe({
          next: (data) => {
            if (!active) {
              return;
            }
            const items: SpotlightItem[] = [];
            for (const edge of data.search.edges) {
              const typename = edge.node.__typename;
              if (typename == null || typename === "%other") {
                continue;
              }
              const to = resourceHref(
                organizationId,
                typename,
                edge.node.id,
                searchHitParentId(edge.node),
              );
              if (to == null) {
                continue;
              }
              items.push({
                id: `resource:${edge.node.id}`,
                kind: "resource",
                label: edge.title,
                groupLabel: t(`nav.spotlight.types.${typename}`, {
                  defaultValue: typename,
                }),
                to,
              });
            }
            setResources(items);
          },
          error: () => {
            setResources([]);
          },
        });
      });
    }, 200);

    return () => {
      active = false;
      window.clearTimeout(handle);
      subscription?.unsubscribe();
    };
  }, [input, open, organizationId, t]);

  const showPages = !hasQualifier(input);
  const items = useMemo(() => {
    if (!open || input.trim() === "") {
      return destinations;
    }
    if (showPages) {
      return [...resources, ...destinations];
    }
    return resources;
  }, [destinations, input, open, resources, showPages]);

  useHotkeys("mod+k", () => {
    handleOpenChange(!open);
  }, { preventDefault: true });

  function onValueChange(item: SpotlightItem | null) {
    if (item == null) {
      return;
    }
    void navigate(item.to);
    handleOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogPopup placement="top" lockScroll>
        <DialogTitle className={slots.title()}>{t("nav.spotlight.title")}</DialogTitle>
        <DialogBody>
          <Combobox<SpotlightItem>
            items={items}
            value={null}
            onValueChange={onValueChange}
            inputValue={input}
            onInputValueChange={setInput}
            itemToStringLabel={item => `${item.label} ${item.groupLabel}`}
            filter={(item, query) => {
              if (item.kind === "resource") {
                return true;
              }
              const haystack = `${item.label} ${item.groupLabel}`.toLowerCase();
              return haystack.includes(query.trim().toLowerCase());
            }}
            inline
            open={open}
            onOpenChange={handleOpenChange}
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
            <ComboboxList<SpotlightItem> className={slots.list()}>
              {item => (
                <ComboboxItem key={item.id} value={item}>
                  <span className={slots.row()}>
                    <span>{item.label}</span>
                    <span className={slots.group()}>{item.groupLabel}</span>
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
