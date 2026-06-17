// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Breadcrumb, PageHeader } from "@probo/ui";
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(trackerPatternDetailPageQuery, queryRef);

  if (data.cookieBanner.__typename !== "CookieBanner") {
    throw new Error("invalid type for cookieBanner node");
  }
  if (data.node.__typename !== "TrackerPattern") {
    throw new Error("invalid type for node");
  }

  const cookieBanner = data.cookieBanner;
  const pattern = data.node;

  return (
    <div className="space-y-6" data-theme="v2">
      <Breadcrumb
        items={[
          {
            label: __("Cookie Banners"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: cookieBanner.name,
            to: `/organizations/${organizationId}/cookie-banners/${cookieBanner.id}/settings`,
          },
          {
            label: __("Trackers"),
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
