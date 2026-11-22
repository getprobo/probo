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

import { usePageTitle } from "@probo/hooks";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { TrackerPatternDetailPageQuery } from "#/__generated__/core/TrackerPatternDetailPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { cookieBannerOverview } from "../overview/variants";

import { TrackerPatternDetailHeader } from "./_components/TrackerPatternDetailHeader";
import { TrackerPatternDetectedTrackersSection } from "./_components/TrackerPatternDetectedTrackersSection";
import { TrackerPatternPropertiesSection } from "./_components/TrackerPatternPropertiesSection";
import { TrackerHitsChart } from "./TrackerHitsChart";

export const trackerPatternDetailPageQuery = graphql`
  query TrackerPatternDetailPageQuery(
    $cookieBannerId: ID!
    $trackerPatternId: ID!
  ) {
    cookieBanner: node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...TrackerPatternPropertiesSection_cookieBanner
      }
    }
    node(id: $trackerPatternId) @required(action: THROW) {
      __typename
      ... on TrackerPattern {
        pattern
        ...TrackerPatternDetailHeader_trackerPattern
        ...TrackerHitsChart_pattern
        ...TrackerPatternPropertiesSection_trackerPattern
        ...TrackerPatternDetectedTrackersSection_trackerPattern
      }
    }
  }
`;

interface TrackerPatternDetailPageProps {
  queryRef: PreloadedQuery<TrackerPatternDetailPageQuery>;
}

export default function TrackerPatternDetailPage({
  queryRef,
}: TrackerPatternDetailPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { card } = cookieBannerOverview();
  const data = usePreloadedQuery<TrackerPatternDetailPageQuery>(trackerPatternDetailPageQuery, queryRef);

  if (data.cookieBanner.__typename !== "CookieBanner") {
    throw new NotFoundError("Cookie banner not found");
  }
  if (data.node.__typename !== "TrackerPattern") {
    throw new NotFoundError("Tracker pattern not found");
  }

  const pattern = data.node;
  usePageTitle(pattern.pattern);

  return (
    <div className="space-y-6">
      <TrackerPatternDetailHeader trackerPatternKey={pattern} />

      <Card size={2} variant="soft">
        <div className={card()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("trackerHits.chart")}
          </Heading>
          <TrackerHitsChart pattern={pattern} />
        </div>
      </Card>

      <TrackerPatternPropertiesSection
        trackerPatternKey={pattern}
        cookieBannerKey={data.cookieBanner}
      />

      <TrackerPatternDetectedTrackersSection trackerPatternKey={pattern} />
    </div>
  );
}
