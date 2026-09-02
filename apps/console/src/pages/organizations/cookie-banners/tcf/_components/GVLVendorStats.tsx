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

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { TFunction } from "i18next";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { GVLVendorStats_cookieBanner$key } from "#/__generated__/core/GVLVendorStats_cookieBanner.graphql";
import type { GVLVendorStats_query$key } from "#/__generated__/core/GVLVendorStats_query.graphql";

import { gvlVendorStats } from "../variants";

const queryFragment = graphql`
  fragment GVLVendorStats_query on Query {
    commonGVLCatalog {
      vendorListVersion
      tcfPolicyVersion
    }
    catalogVendors: commonGVLVendors(first: 1) {
      totalCount
    }
  }
`;

const cookieBannerFragment = graphql`
  fragment GVLVendorStats_cookieBanner on CookieBanner {
    gvlVendors(first: 500) {
      totalCount
      edges {
        node {
          iabVendorId
        }
      }
    }
    publishedVersion {
      gvlVendorCount
      gvlVendorIds
    }
  }
`;

interface GVLVendorStatsProps {
  queryKey: GVLVendorStats_query$key;
  cookieBannerKey: GVLVendorStats_cookieBanner$key;
}

export function GVLVendorStats({ queryKey, cookieBannerKey }: GVLVendorStatsProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const query = useFragment<GVLVendorStats_query$key>(queryFragment, queryKey);
  const banner = useFragment<GVLVendorStats_cookieBanner$key>(
    cookieBannerFragment,
    cookieBannerKey,
  );
  const { root } = gvlVendorStats();
  const missing = t("tcfPage.stats.empty");
  const draftCount = banner.gvlVendors?.totalCount ?? 0;
  const liveCount = banner.publishedVersion?.gvlVendorCount;
  const status = vendorStatus(
    (banner.gvlVendors?.edges ?? []).map(edge => edge.node.iabVendorId),
    banner.publishedVersion?.gvlVendorIds,
  );

  return (
    <div className={root()}>
      <StatTile
        title={t("tcfPage.stats.vendors")}
        entries={[
          { key: "draft", value: String(draftCount), unit: t("tcfPage.stats.draft") },
          {
            key: "live",
            value: liveCount == null ? missing : String(liveCount),
            unit: t("tcfPage.stats.live"),
          },
        ]}
      >
        {status.kind === "never" && (
          <Badge
            size={1}
            variant="soft"
            color="amber"
          >
            {t("tcfPage.stats.unpublishedNever")}
          </Badge>
        )}
        {status.kind === "added" && (
          <Badge
            size={1}
            variant="soft"
            color="amber"
          >
            {t("tcfPage.stats.unpublishedAdded", { count: status.count })}
          </Badge>
        )}
        {status.kind === "removed" && (
          <Badge
            size={1}
            variant="soft"
            color="amber"
          >
            {t("tcfPage.stats.unpublishedRemoved", { count: status.count })}
          </Badge>
        )}
        {status.kind === "changed" && (
          <Badge
            size={1}
            variant="soft"
            color="amber"
          >
            {t("tcfPage.stats.unpublishedChanged", {
              added: status.added,
              removed: status.removed,
            })}
          </Badge>
        )}
      </StatTile>
      <StatTile
        title={t("tcfPage.stats.catalog")}
        entries={[
          {
            key: "gvl",
            value: formatVersion(query.commonGVLCatalog.vendorListVersion, missing, t),
            unit: t("tcfPage.stats.gvl"),
          },
          {
            key: "tcfPolicy",
            value: formatVersion(query.commonGVLCatalog.tcfPolicyVersion, missing, t),
            unit: t("tcfPage.stats.tcfPolicy"),
          },
        ]}
      >
        <Badge
          size={1}
          variant="soft"
          color="sky"
        >
          {t("tcfPage.stats.catalogTotal", { count: query.catalogVendors.totalCount })}
        </Badge>
      </StatTile>
    </div>
  );
}

interface StatTileProps {
  title: string;
  entries: readonly { key: string; value: string; unit: string }[];
  children?: ReactNode;
}

function StatTile({ title, entries, children }: StatTileProps) {
  const { card, comparison, valueRow, value, footer } = gvlVendorStats();

  return (
    <Card
      variant="soft"
      size={3}
      padding={3}
      className={card()}
    >
      <Text size={2} weight="medium">
        {title}
      </Text>
      <div className={comparison()}>
        {entries.map(entry => (
          <div key={entry.key} className={valueRow()}>
            <Text size={6} weight="bold" highContrast className={value()}>
              {entry.value}
            </Text>
            <Text size={2}>
              {entry.unit}
            </Text>
          </div>
        ))}
      </div>
      {children && <div className={footer()}>{children}</div>}
    </Card>
  );
}

type VendorStatus =
  | { kind: "synced" }
  | { kind: "never" }
  | { kind: "added"; count: number }
  | { kind: "removed"; count: number }
  | { kind: "changed"; added: number; removed: number };

/**
 * Compares the draft and published vendor sets by id rather than by size, so a
 * swap of one vendor for another is reported as a pending change instead of
 * reading as synced.
 */
function vendorStatus(
  draftIDs: readonly number[],
  liveIDs: readonly number[] | null | undefined,
): VendorStatus {
  if (liveIDs == null) {
    return { kind: "never" };
  }

  const live = new Set(liveIDs);
  const draft = new Set(draftIDs);
  const added = draftIDs.filter(id => !live.has(id)).length;
  const removed = liveIDs.filter(id => !draft.has(id)).length;

  if (added === 0 && removed === 0) {
    return { kind: "synced" };
  }

  if (removed === 0) {
    return { kind: "added", count: added };
  }

  if (added === 0) {
    return { kind: "removed", count: removed };
  }

  return { kind: "changed", added, removed };
}

function formatVersion(
  value: number | null | undefined,
  missing: string,
  t: TFunction,
): string {
  if (value == null) {
    return missing;
  }

  return t("tcfPage.stats.version", { version: value });
}
