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

import { Breadcrumb, PageHeader } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { TrackerPatternDetailPageQuery } from "#/__generated__/core/TrackerPatternDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { TrackerPatternDetectedTrackersSection } from "./_components/TrackerPatternDetectedTrackersSection";
import { TrackerPatternPropertiesSection } from "./_components/TrackerPatternPropertiesSection";

export const trackerPatternDetailPageQuery = graphql`
  query TrackerPatternDetailPageQuery(
    $cookieBannerId: ID!
    $trackerPatternId: ID!
  ) {
    cookieBanner: node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        id
        name
      }
    }
    node(id: $trackerPatternId) @required(action: THROW) {
      __typename
      ... on TrackerPattern {
        id
        displayName
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
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<TrackerPatternDetailPageQuery>(trackerPatternDetailPageQuery, queryRef);

  if (data.cookieBanner.__typename !== "CookieBanner") {
    throw new Error("invalid type for cookieBanner node");
  }
  if (data.node.__typename !== "TrackerPattern") {
    throw new Error("invalid type for node");
  }

  const cookieBanner = data.cookieBanner;
  const pattern = data.node;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("trackerPatternDetailPage.breadcrumbs.index"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: cookieBanner.name,
            to: `/organizations/${organizationId}/cookie-banners/${cookieBanner.id}/settings`,
          },
          {
            label: t("trackerPatternDetailPage.breadcrumbs.trackers"),
            to: `/organizations/${organizationId}/cookie-banners/${cookieBanner.id}/trackers`,
          },
          {
            label: pattern.displayName,
          },
        ]}
      />

      <PageHeader title={pattern.displayName} />

      <TrackerPatternPropertiesSection trackerPatternKey={pattern} />

      <TrackerPatternDetectedTrackersSection trackerPatternKey={pattern} />
    </div>
  );
}
