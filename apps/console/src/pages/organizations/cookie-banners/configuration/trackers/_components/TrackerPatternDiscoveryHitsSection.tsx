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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrackerPatternDiscoveryHitsSection_cookieBanner$key } from "#/__generated__/core/TrackerPatternDiscoveryHitsSection_cookieBanner.graphql";
import type { TrackerPatternDiscoveryHitsSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternDiscoveryHitsSection_trackerPattern.graphql";

import { trackerPatternDiscoveryHitsSection } from "../../../variants";
import { DiscoveryFamilyCounts } from "../../_components/DiscoveryFamilyCounts";

const cookieBannerFragment = graphql`
  fragment TrackerPatternDiscoveryHitsSection_cookieBanner on CookieBanner {
    discoveryPageLoads {
      family
      count
    }
  }
`;

const trackerPatternFragment = graphql`
  fragment TrackerPatternDiscoveryHitsSection_trackerPattern on TrackerPattern {
    discoveryHits {
      family
      count
    }
  }
`;

interface TrackerPatternDiscoveryHitsSectionProps {
  trackerPatternKey: TrackerPatternDiscoveryHitsSection_trackerPattern$key;
  cookieBannerKey: TrackerPatternDiscoveryHitsSection_cookieBanner$key;
}

export function TrackerPatternDiscoveryHitsSection({
  trackerPatternKey,
  cookieBannerKey,
}: TrackerPatternDiscoveryHitsSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(cookieBannerFragment, cookieBannerKey);
  const pattern = useFragment(trackerPatternFragment, trackerPatternKey);
  const { root, intro } = trackerPatternDiscoveryHitsSection();

  if (pattern.discoveryHits.length === 0) {
    return null;
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={intro()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("trackerDiscovery.title")}
          </Heading>
          <Text size={2} color="neutral">
            {t("trackerDiscovery.description")}
          </Text>
        </div>
        <DiscoveryFamilyCounts
          items={pattern.discoveryHits}
          denominators={banner.discoveryPageLoads}
        />
      </div>
    </Card>
  );
}
