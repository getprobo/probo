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
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrackerSourceBadge_trackerPattern$key } from "#/__generated__/core/TrackerSourceBadge_trackerPattern.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { cookieBannerTrackersPath } from "../../../_lib/cookieBannerPaths";
import { trackerSourceBadge } from "../../../variants";
import { cookieSourceBadges } from "../_lib/trackerBadges";

const trackerSourceBadgeFragment = graphql`
  fragment TrackerSourceBadge_trackerPattern on TrackerPattern {
    pattern
    source
    sourceCookieBanner {
      id
      origin
    }
  }
`;

export function originHost(origin: string): string {
  try {
    return new URL(origin).host;
  } catch {
    return origin;
  }
}

interface TrackerSourceBadgeProps {
  trackerPatternKey: TrackerSourceBadge_trackerPattern$key;
  layout: "stacked" | "inline";
  empty?: "dash" | "hidden";
}

export function TrackerSourceBadge({
  trackerPatternKey,
  layout,
  empty = "dash",
}: TrackerSourceBadgeProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const pattern = useFragment(trackerSourceBadgeFragment, trackerPatternKey);
  const { root, via } = trackerSourceBadge({ layout });
  const sourceBadge = pattern.source == null ? null : cookieSourceBadges[pattern.source];
  const sourceBanner = pattern.sourceCookieBanner;

  if (sourceBadge == null) {
    if (empty === "hidden") {
      return null;
    }
    return <Text size={2} color="faint">-</Text>;
  }

  const search = new URLSearchParams();
  if (sourceBanner != null) {
    search.set("view", "all");
    search.set("q", pattern.pattern);
  }

  return (
    <div className={root()}>
      <Badge variant={sourceBadge.variant} color={sourceBadge.color}>
        {t(`trackerPatternRow.sources.${sourceBadge.labelKey}`)}
      </Badge>
      {sourceBanner != null && (
        <Link
          to={{
            pathname: cookieBannerTrackersPath(organizationId, sourceBanner.id),
            search: search.toString(),
          }}
          size={1}
          color="neutral"
          underline={false}
          className={via()}
        >
          {t("trackerPatternRow.sourceVia", { origin: originHost(sourceBanner.origin) })}
        </Link>
      )}
    </div>
  );
}
