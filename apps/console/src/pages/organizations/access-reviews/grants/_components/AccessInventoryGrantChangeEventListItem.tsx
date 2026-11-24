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
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessInventoryGrantChangeEventListItem_event$key } from "#/__generated__/core/AccessInventoryGrantChangeEventListItem_event.graphql";

import { accessInventoryGrantChangeEventListItem } from "./variants";

const accessInventoryGrantChangeEventListItemFragment = graphql`
  fragment AccessInventoryGrantChangeEventListItem_event on AccessInventoryChangeEvent {
    id
    changeType
    entitlementKey
    effectiveAt
    observedAt
    providerOccurredAt
    before {
      entitlementDisplayName
      presence
    }
    after {
      entitlementDisplayName
      presence
    }
  }
`;

interface AccessInventoryGrantChangeEventListItemProps {
  eventKey: AccessInventoryGrantChangeEventListItem_event$key;
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

function presenceLabel(
  presence: "PRESENT" | "DISAPPEARED",
  t: (key: string) => string,
): string {
  return t(
    presence === "DISAPPEARED"
      ? "grants.presence.disappeared"
      : "grants.presence.present",
  );
}

export function AccessInventoryGrantChangeEventListItem({
  eventKey,
}: AccessInventoryGrantChangeEventListItemProps) {
  const { i18n, t } = useTranslation("organizations/access-reviews");
  const event = useFragment(
    accessInventoryGrantChangeEventListItemFragment,
    eventKey,
  );
  const {
    root,
    header,
    meta,
    snapshots,
    snapshot,
    snapshotTitle,
    snapshotBody,
  } = accessInventoryGrantChangeEventListItem();

  const showSnapshots = event.before != null || event.after != null;
  const entitlementFallback = event.entitlementKey.trim() !== ""
    ? event.entitlementKey
    : t("grants.values.unknown");

  return (
    <div className={root()}>
      <div className={header()}>
        <Badge
          size={1}
          variant="soft"
          color={changeTypeBadgeColor(event.changeType)}
        >
          {t(`grantDetail.changeTypes.${event.changeType}`)}
        </Badge>
        <Text size={2} highContrast>
          {dateTimeFormat(i18n.language, event.effectiveAt)}
        </Text>
      </div>
      <div className={meta()}>
        <Text size={1} color="faint">
          {t("grantDetail.changelog.observedAt", {
            value: dateTimeFormat(i18n.language, event.observedAt),
          })}
        </Text>
        {event.providerOccurredAt != null && (
          <Text size={1} color="faint">
            {t("grantDetail.changelog.providerOccurredAt", {
              value: dateTimeFormat(i18n.language, event.providerOccurredAt),
            })}
          </Text>
        )}
      </div>
      {showSnapshots && (
        <div className={snapshots()}>
          <div className={snapshot()}>
            <Text size={1} className={snapshotTitle()}>
              {t("grantDetail.changelog.before")}
            </Text>
            {event.before == null
              ? (
                  <Text size={2} color="faint">
                    {t("grantDetail.values.empty")}
                  </Text>
                )
              : (
                  <div className={snapshotBody()}>
                    <Text size={2} highContrast>
                      {event.before.entitlementDisplayName.trim()
                        || entitlementFallback}
                    </Text>
                    <Text size={1} color="faint">
                      {presenceLabel(event.before.presence, t)}
                    </Text>
                  </div>
                )}
          </div>
          <div className={snapshot()}>
            <Text size={1} className={snapshotTitle()}>
              {t("grantDetail.changelog.after")}
            </Text>
            {event.after == null
              ? (
                  <Text size={2} color="faint">
                    {t("grantDetail.values.empty")}
                  </Text>
                )
              : (
                  <div className={snapshotBody()}>
                    <Text size={2} highContrast>
                      {event.after.entitlementDisplayName.trim()
                        || entitlementFallback}
                    </Text>
                    <Text size={1} color="faint">
                      {presenceLabel(event.after.presence, t)}
                    </Text>
                  </div>
                )}
          </div>
        </div>
      )}
    </div>
  );
}
