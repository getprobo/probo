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

import { PencilSimpleIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type MouseEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalAccessListItemFragment$key } from "#/__generated__/core/CompliancePortalAccessListItemFragment.graphql";

import { accessListItem } from "../variants";

import { CompliancePortalAccessEditDialog } from "./CompliancePortalAccessEditDialog";

const fragment = graphql`
  fragment CompliancePortalAccessListItemFragment on CompliancePortalAccess {
    id
    createdAt
    profile {
      fullName
      emailAddress
      state
    }
    activeCount
    pendingRequestCount
    ndaSignature {
      status
    }
    canUpdate: permission(action: "compliance-portal:portal-access:update")
  }
`;

type ElectronicSignatureStatus = "PENDING" | "ACCEPTED" | "PROCESSING" | "COMPLETED" | "FAILED";

function ndaBadgeColor(
  status: ElectronicSignatureStatus,
): "green" | "sky" | "amber" | "red" {
  switch (status) {
    case "COMPLETED":
      return "green";
    case "ACCEPTED":
    case "PROCESSING":
      return "sky";
    case "PENDING":
      return "amber";
    case "FAILED":
      return "red";
  }
}

function ndaBadgeKey(status: ElectronicSignatureStatus): string {
  switch (status) {
    case "COMPLETED":
      return "ndaSignatureBadge.signed";
    case "ACCEPTED":
    case "PROCESSING":
      return "ndaSignatureBadge.processing";
    case "PENDING":
      return "ndaSignatureBadge.pending";
    case "FAILED":
      return "ndaSignatureBadge.failed";
  }
}

interface CompliancePortalAccessListItemProps {
  accessKey: CompliancePortalAccessListItemFragment$key;
}

export function CompliancePortalAccessListItem({
  accessKey,
}: CompliancePortalAccessListItemProps) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);

  const access = useFragment<CompliancePortalAccessListItemFragment$key>(fragment, accessKey);

  const isActive = access.profile.state === "ACTIVE";
  const canEdit = access.canUpdate && isActive;
  const { item, trailing, counts } = accessListItem({
    interactive: canEdit,
    inactive: !isActive,
  });

  function handleRowClick() {
    if (canEdit) {
      setDialogOpen(true);
    }
  }

  function handleEditClick(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation();
    setDialogOpen(true);
  }

  return (
    <>
      <ListItem key={access.id} className={item()} onClick={handleRowClick}>
        <Avatar
          size={2}
          variant="soft"
          color="gold"
          fallback={access.profile.fullName.charAt(0).toUpperCase() || "?"}
        />
        <ListItemContent>
          <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
            {access.profile.fullName}
          </Text>
          <Text size={1} color="gold" className="truncate">
            {access.profile.emailAddress}
          </Text>
        </ListItemContent>
        <div className={trailing()}>
          <Text size={1} color="faint">
            {dateFormat(i18n.language, access.createdAt)}
          </Text>
          <div className={counts()}>
            <Text size={1} color="faint">
              {t("accessListItem.granted", { count: access.activeCount })}
            </Text>
            {access.pendingRequestCount > 0 && (
              <Text size={1} color="faint">
                {t("accessListItem.pending", { count: access.pendingRequestCount })}
              </Text>
            )}
          </div>
          {access.ndaSignature
            ? (
                <Badge color={ndaBadgeColor(access.ndaSignature.status)} variant="soft">
                  {t(ndaBadgeKey(access.ndaSignature.status))}
                </Badge>
              )
            : null}
          {canEdit && (
            <IconButton
              variant="ghost"
              color="neutral"
              aria-label={t("accessListItem.actions.edit")}
              onClick={handleEditClick}
            >
              <PencilSimpleIcon />
            </IconButton>
          )}
        </div>
      </ListItem>

      {canEdit && dialogOpen && (
        <CompliancePortalAccessEditDialog
          access={access}
          onClose={() => setDialogOpen(false)}
        />
      )}
    </>
  );
}
