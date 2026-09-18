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

import { MinusIcon, PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { GVLVendorListItem_commonGVLVendor$key } from "#/__generated__/core/GVLVendorListItem_commonGVLVendor.graphql";

import { gvlVendorListItem } from "../variants";

const gvlVendorListItemFragment = graphql`
  fragment GVLVendorListItem_commonGVLVendor on CommonGVLVendor {
    iabVendorId
    name
    policyUrl
  }
`;

interface GVLVendorListItemProps {
  commonGVLVendorKey: GVLVendorListItem_commonGVLVendor$key;
  selected: boolean;
  onPress?: () => void;
}

export function GVLVendorListItem({
  commonGVLVendorKey,
  selected,
  onPress,
}: GVLVendorListItemProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const vendor = useFragment(gvlVendorListItemFragment, commonGVLVendorKey);
  const { name, meta, trailing } = gvlVendorListItem();
  const canUpdate = onPress != null;
  const actionLabel = selected
    ? t("tcfPage.actions.removeVendor", { name: vendor.name })
    : t("tcfPage.actions.addVendor", { name: vendor.name });

  const action = (
    <Button
      type="button"
      size={2}
      variant="ghost"
      color={selected ? "red" : "neutral"}
      disabled={!canUpdate}
      aria-label={actionLabel}
      iconStart={selected ? <MinusIcon /> : <PlusIcon />}
      onClick={onPress}
    >
      {selected ? t("tcfPage.actions.remove") : t("tcfPage.actions.add")}
    </Button>
  );

  return (
    <ListItem>
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast className={name()}>
          {vendor.name}
        </Text>
        <Text size={1} color="faint" className={meta()}>
          {t("tcfPage.iabVendorId", { id: vendor.iabVendorId })}
          {vendor.policyUrl
            ? (
                <>
                  {" · "}
                  <Anchor
                    href={vendor.policyUrl}
                    target="_blank"
                    rel="noreferrer"
                    size={1}
                  >
                    {t("tcfPage.policy")}
                  </Anchor>
                </>
              )
            : null}
        </Text>
      </ListItemContent>
      <div className={trailing()}>
        {canUpdate
          ? action
          : (
              <Tooltip>
                <TooltipTrigger render={<span tabIndex={0}>{action}</span>} />
                <TooltipPopup>{t("tcfPage.noUpdatePermission")}</TooltipPopup>
              </Tooltip>
            )}
      </div>
    </ListItem>
  );
}
