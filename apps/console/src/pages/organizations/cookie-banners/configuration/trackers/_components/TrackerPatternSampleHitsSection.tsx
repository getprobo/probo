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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrackerPatternSampleHitsSection_cookieBanner$key } from "#/__generated__/core/TrackerPatternSampleHitsSection_cookieBanner.graphql";
import type { TrackerPatternSampleHitsSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternSampleHitsSection_trackerPattern.graphql";

import { SampleFamilyCounts } from "../../_components/SampleFamilyCounts";

const cookieBannerFragment = graphql`
  fragment TrackerPatternSampleHitsSection_cookieBanner on CookieBanner {
    latestSample {
      pageLoads {
        family
        count
      }
      expectedPageLoads {
        family
        count
      }
    }
  }
`;

const trackerPatternFragment = graphql`
  fragment TrackerPatternSampleHitsSection_trackerPattern on TrackerPattern {
    sampleHits {
      family
      count
    }
  }
`;

interface TrackerPatternSampleHitsSectionProps {
  trackerPatternKey: TrackerPatternSampleHitsSection_trackerPattern$key;
  cookieBannerKey: TrackerPatternSampleHitsSection_cookieBanner$key;
}

export function TrackerPatternSampleHitsSection({
  trackerPatternKey,
  cookieBannerKey,
}: TrackerPatternSampleHitsSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(cookieBannerFragment, cookieBannerKey);
  const pattern = useFragment(trackerPatternFragment, trackerPatternKey);

  if (pattern.sampleHits.length === 0) {
    return (
      <Text size={2} color="faint">
        {t("trackerSample.empty")}
      </Text>
    );
  }

  return (
    <SampleFamilyCounts
      items={pattern.sampleHits.map(hit => ({
        family: hit.family,
        count: hit.count,
      }))}
      denominators={(banner.latestSample?.pageLoads ?? []).map(load => ({
        family: load.family,
        count: load.count,
      }))}
      expectedDenominators={(banner.latestSample?.expectedPageLoads ?? []).map(load => ({
        family: load.family,
        count: load.count,
      }))}
    />
  );
}
