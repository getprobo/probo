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

import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableLink } from "@probo/ui/src/v2/Table/TableLink";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { AccessInventoryGrantListItem_grant$key } from "#/__generated__/core/AccessInventoryGrantListItem_grant.graphql";

import { accessReviewSourceLabel } from "../../_lib/sourceLabel";

import { accessInventoryGrantsList } from "./variants";

const accessInventoryGrantListItemFragment = graphql`
  fragment AccessInventoryGrantListItem_grant on AccessInventoryGrant {
    id
    entitlementDisplayName
    entitlementKey
    presence
    lastSeenAt
    principal {
      email
      fullName
    }
    resource {
      displayName
      kind
    }
    source {
      name
      connectorAccount {
        name
      }
    }
  }
`;

interface AccessInventoryGrantListItemProps {
  grantKey: AccessInventoryGrantListItem_grant$key;
}

export function AccessInventoryGrantListItem({
  grantKey,
}: AccessInventoryGrantListItemProps) {
  const { i18n, t } = useTranslation("organizations/access-reviews");
  const location = useLocation();
  const grant = useFragment(accessInventoryGrantListItemFragment, grantKey);
  const disappeared = grant.presence === "DISAPPEARED";
  const { principal, resource } = accessInventoryGrantsList({ disappeared });
  const fullName = grant.principal.fullName.trim();
  const email = grant.principal.email.trim();
  const principalTitle = fullName || email || t("grants.values.unknown");
  const entitlement = grant.entitlementDisplayName.trim()
    || grant.entitlementKey
    || t("grants.values.unknown");
  const resourceName = grant.resource.displayName.trim()
    || t("grants.values.unknown");
  const sourceLabel = accessReviewSourceLabel({
    name: grant.source.name,
    accountName: grant.source.connectorAccount?.name,
  });

  return (
    <TableRow interactive>
      <TableRowHeaderCell overflow="truncate">
        <TableLink
          to={{ pathname: grant.id, search: location.search }}
          aria-label={t("grantDetail.open", { name: principalTitle })}
        >
          <div className={principal()}>
            <Text size={2} weight="medium" highContrast className="truncate">
              {principalTitle}
            </Text>
            {fullName !== "" && email !== "" && (
              <Text size={1} color="faint" className="truncate">
                {email}
              </Text>
            )}
          </div>
        </TableLink>
      </TableRowHeaderCell>
      <TableCell overflow="truncate">
        <div className={resource()}>
          <Text size={2} highContrast className="truncate">
            {resourceName}
          </Text>
          {grant.resource.kind.trim() !== "" && (
            <Text size={1} color="faint" className="truncate">
              {grant.resource.kind}
            </Text>
          )}
        </div>
      </TableCell>
      <TableCell overflow="truncate">
        <Text size={2} className="truncate">
          {entitlement}
        </Text>
      </TableCell>
      <TableCell overflow="truncate">
        <Text size={2} className="truncate">
          {sourceLabel}
        </Text>
      </TableCell>
      <TableCell>
        <Badge
          size={1}
          variant="soft"
          color={disappeared ? "amber" : "green"}
        >
          {t(
            disappeared
              ? "grants.presence.disappeared"
              : "grants.presence.present",
          )}
        </Badge>
      </TableCell>
      <TableCell>
        <Text size={2} color="neutral" className="whitespace-nowrap">
          {dateTimeFormat(i18n.language, grant.lastSeenAt)}
        </Text>
      </TableCell>
    </TableRow>
  );
}
