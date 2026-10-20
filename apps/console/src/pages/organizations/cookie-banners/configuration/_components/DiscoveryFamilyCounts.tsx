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

import { discoveryFamilyCounts } from "../../variants";
import { familyLabel } from "../_lib/familyLabel";

import { FamilyLogo } from "./FamilyLogo";

export interface DiscoveryFamilyCountItem {
  family: string;
  count: number;
}

export interface DiscoveryFamilyCountsProps {
  items: readonly DiscoveryFamilyCountItem[];
  denominators?: readonly DiscoveryFamilyCountItem[];
}

export function DiscoveryFamilyCounts({
  items,
  denominators,
}: DiscoveryFamilyCountsProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const { root, card, family, logo, stats, stat, value }
    = discoveryFamilyCounts();
  const loadsByFamily = new Map(
    (denominators ?? []).map(entry => [entry.family, entry.count]),
  );

  return (
    <div className={root()}>
      {items.map((entry) => {
        const loads = loadsByFamily.get(entry.family);
        const rate = loads != null && loads > 0
          ? formatHitRate(entry.count, loads, i18n.language)
          : null;

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
            {rate != null
              ? (
                  <div className={stat()}>
                    <Text size={4} weight="bold" highContrast className={value()}>
                      {rate}
                    </Text>
                    <Text size={2}>{t("trackerDiscovery.hitRate")}</Text>
                  </div>
                )
              : null}
            <div className={stats()}>
              <div className={stat()}>
                <Text size={4} weight="bold" highContrast className={value()}>
                  {entry.count}
                </Text>
                <Text size={2}>{t("trackerDiscovery.hits")}</Text>
              </div>
              {loads != null
                ? (
                    <>
                      <Text size={4} color="faint">/</Text>
                      <div className={stat()}>
                        <Text size={4} weight="bold" highContrast className={value()}>
                          {loads}
                        </Text>
                        <Text size={2}>{t("trackerDiscovery.loads")}</Text>
                      </div>
                    </>
                  )
                : null}
            </div>
          </Card>
        );
      })}
    </div>
  );
}

function formatHitRate(hits: number, loads: number, locale: string): string {
  return new Intl.NumberFormat(locale, {
    style: "percent",
    maximumFractionDigits: 0,
  }).format(hits / loads);
}
