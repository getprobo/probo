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

import { sampleFamilyCounts } from "../../variants";
import { familyLabel } from "../_lib/familyLabel";

import { FamilyLogo } from "./FamilyLogo";

export interface SampleFamilyCountItem {
  family: string;
  count: number;
}

export interface SampleFamilyCountsProps {
  items: readonly SampleFamilyCountItem[];
  denominators?: readonly SampleFamilyCountItem[];
  allowedHits?: readonly SampleFamilyCountItem[];
  deniedHits?: readonly SampleFamilyCountItem[];
}

export function SampleFamilyCounts({
  items,
  denominators,
  allowedHits,
  deniedHits,
}: SampleFamilyCountsProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const { root, card, family, logo, rates, rate, rateLead, value, fraction }
    = sampleFamilyCounts();
  const loadsByFamily = new Map(
    (denominators ?? []).map(entry => [entry.family, entry.count]),
  );
  const allowedHitsByFamily = new Map(
    (allowedHits ?? []).map(entry => [entry.family, entry.count]),
  );
  const deniedHitsByFamily = new Map(
    (deniedHits ?? []).map(entry => [entry.family, entry.count]),
  );

  return (
    <div className={root()}>
      {items.map((entry) => {
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
