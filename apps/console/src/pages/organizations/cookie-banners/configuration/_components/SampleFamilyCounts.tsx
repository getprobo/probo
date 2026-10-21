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
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SampleFamilyCounts_cookieBanner$key } from "#/__generated__/core/SampleFamilyCounts_cookieBanner.graphql";
import type { SampleFamilyCounts_trackerPattern$key } from "#/__generated__/core/SampleFamilyCounts_trackerPattern.graphql";

import { sampleFamilyCounts } from "../../variants";
import { familyLabel } from "../_lib/familyLabel";

import { FamilyLogo } from "./FamilyLogo";

const cookieBannerFragment = graphql`
  fragment SampleFamilyCounts_cookieBanner on CookieBanner {
    latestSample {
      pageLoads {
        family
        count
      }
    }
  }
`;

const trackerPatternFragment = graphql`
  fragment SampleFamilyCounts_trackerPattern on TrackerPattern {
    sampleHits {
      family
      count
    }
    allowedSampleHits {
      family
      count
    }
    deniedSampleHits {
      family
      count
    }
  }
`;

interface SampleFamilyCountsProps {
  cookieBannerKey: SampleFamilyCounts_cookieBanner$key;
  trackerPatternKey: SampleFamilyCounts_trackerPattern$key;
}

export function SampleFamilyCounts({
  cookieBannerKey,
  trackerPatternKey,
}: SampleFamilyCountsProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(cookieBannerFragment, cookieBannerKey);
  const pattern = useFragment(trackerPatternFragment, trackerPatternKey);
  const { root, card, family, logo, rates, rate, rateLead, value, fraction }
    = sampleFamilyCounts();
  const loadsByFamily = new Map(
    (banner.latestSample?.pageLoads ?? []).map(entry => [entry.family, entry.count]),
  );
  const allowedHitsByFamily = new Map(
    pattern.allowedSampleHits.map(entry => [entry.family, entry.count]),
  );
  const deniedHitsByFamily = new Map(
    pattern.deniedSampleHits.map(entry => [entry.family, entry.count]),
  );

  if (pattern.sampleHits.length === 0) {
    return (
      <Text size={2} color="faint">
        {t("trackerSample.empty")}
      </Text>
    );
  }

  return (
    <div className={root()}>
      {pattern.sampleHits.map((entry) => {
        const loads = loadsByFamily.get(entry.family) ?? 0;
        const allowedHitCount = allowedHitsByFamily.get(entry.family) ?? 0;
        const deniedHitCount = deniedHitsByFamily.get(entry.family) ?? 0;

        return (
          <Card
            key={entry.family}
            variant="soft"
            size={3}
            padding={4}
            className={card()}
          >
            <div className={family()}>
              <FamilyLogo
                family={entry.family}
                size={24}
                className={logo()}
              />
              <Text size={4} weight="medium">{familyLabel(entry.family, t)}</Text>
            </div>
            <div className={rates()}>
              <RateRow
                label={t("trackerSample.hitRate")}
                hits={entry.count}
                loads={loads}
                locale={i18n.language}
                rateClass={rate()}
                rateLeadClass={rateLead()}
                valueClass={value()}
                fractionClass={fraction()}
              />
              <RateRow
                label={t("trackerSample.expectedHitRate")}
                hits={allowedHitCount}
                loads={loads}
                locale={i18n.language}
                rateClass={rate()}
                rateLeadClass={rateLead()}
                valueClass={value()}
                fractionClass={fraction()}
              />
              <RateRow
                label={t("trackerSample.driftRate")}
                hits={deniedHitCount}
                loads={loads}
                locale={i18n.language}
                rateClass={rate()}
                rateLeadClass={rateLead()}
                valueClass={value()}
                fractionClass={fraction()}
              />
            </div>
          </Card>
        );
      })}
    </div>
  );
}

interface RateRowProps {
  label: string;
  hits: number;
  loads: number;
  locale: string;
  rateClass: string;
  rateLeadClass: string;
  valueClass: string;
  fractionClass: string;
}

function RateRow({
  label,
  hits,
  loads,
  locale,
  rateClass,
  rateLeadClass,
  valueClass,
  fractionClass,
}: RateRowProps) {
  const rate = loads > 0 ? formatHitRate(hits, loads, locale) : "—";

  return (
    <div className={rateClass}>
      <div className={rateLeadClass}>
        <Text size={4} weight="bold" highContrast className={valueClass}>
          {rate}
        </Text>
        <Text size={2}>{label}</Text>
      </div>
      <Text size={2} color="faint" className={fractionClass}>
        {hits}
        {" / "}
        {loads}
      </Text>
    </div>
  );
}

function formatHitRate(hits: number, loads: number, locale: string): string {
  return new Intl.NumberFormat(locale, {
    style: "percent",
    maximumFractionDigits: 0,
  }).format(hits / loads);
}
