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
import { graphql } from "relay-runtime";

import type { AccessInventoryChangeEventListItem_event$key } from "#/__generated__/core/AccessInventoryChangeEventListItem_event.graphql";

import { accessReviewSourceLabel } from "../../_lib/sourceLabel";

import { accessInventoryChangeEventsList } from "./variants";

const accessInventoryChangeEventListItemFragment = graphql`
  fragment AccessInventoryChangeEventListItem_event on AccessInventoryChangeEvent {
    id
    changeType
    principalKey
    resourceKey
    entitlementKey
    effectiveAt
    grant {
      id
      entitlementDisplayName
      principal {
        email
        fullName
      }
      resource {
        displayName
        kind
      }
    }
    source {
      name
      connectorAccount {
        name
      }
    }
  }
`;

interface AccessInventoryChangeEventListItemProps {
  eventKey: AccessInventoryChangeEventListItem_event$key;
}

function changeTypeBadgeColor(
  changeType: "APPEARED" | "DISAPPEARED" | "ATTRIBUTE_CHANGED",
): "green" | "amber" | "sky" {
  if (changeType === "APPEARED") {
    return "green";
  }
  if (changeType === "DISAPPEARED") {
    return "amber";
  }
  return "sky";
}

export function AccessInventoryChangeEventListItem({
  eventKey,
}: AccessInventoryChangeEventListItemProps) {
  const { i18n, t } = useTranslation("organizations/access-reviews");
  const event = useFragment(
    accessInventoryChangeEventListItemFragment,
    eventKey,
  );
  const { principal, resource } = accessInventoryChangeEventsList();

  const grant = event.grant;
  const fullName = grant?.principal.fullName.trim() ?? "";
  const email = grant?.principal.email.trim() ?? "";
  const principalTitle = fullName
    || email
    || event.principalKey
    || t("changes.values.unknown");
  const resourceName = grant?.resource.displayName.trim()
    || event.resourceKey
    || t("changes.values.unknown");
  const resourceKind = grant?.resource.kind.trim() ?? "";
  const entitlement = grant?.entitlementDisplayName.trim()
    || event.entitlementKey
    || t("changes.values.unknown");
  const sourceLabel = accessReviewSourceLabel({
    name: event.source.name,
    accountName: event.source.connectorAccount?.name,
  });

  const principalContent = (
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
  );

  return (
    <TableRow interactive={grant != null}>
      <TableCell>
        <Badge
          size={1}
          variant="soft"
          color={changeTypeBadgeColor(event.changeType)}
        >
          {t(`changes.changeTypes.${event.changeType}`)}
        </Badge>
      </TableCell>
      <TableRowHeaderCell overflow="truncate">
        {grant != null
          ? (
              <TableLink
                to={`../grants/${grant.id}`}
                aria-label={t("changes.openGrant", { name: principalTitle })}
              >
                {principalContent}
              </TableLink>
            )
          : principalContent}
      </TableRowHeaderCell>
      <TableCell overflow="truncate">
        <div className={resource()}>
          <Text size={2} highContrast className="truncate">
            {resourceName}
          </Text>
          {resourceKind !== "" && (
            <Text size={1} color="faint" className="truncate">
              {resourceKind}
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
        <Text size={2} color="neutral" className="whitespace-nowrap">
          {dateTimeFormat(i18n.language, event.effectiveAt)}
        </Text>
      </TableCell>
    </TableRow>
  );
}
