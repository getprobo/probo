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
import { Fragment } from "react";
import { useTranslation } from "react-i18next";

import { discoveryFamilyCounts } from "../../variants";
import { familyLabel } from "../_lib/familyLabel";

export interface DiscoveryFamilyCountItem {
  family: string;
  count: number;
}

export interface DiscoveryFamilyCountsProps {
  items: readonly DiscoveryFamilyCountItem[];
  denominators?: readonly DiscoveryFamilyCountItem[];
  compact?: boolean;
}

export function DiscoveryFamilyCounts({
  items,
  denominators,
  compact = false,
}: DiscoveryFamilyCountsProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { root, item, separator } = discoveryFamilyCounts({ compact });
  const loadsByFamily = new Map(
    (denominators ?? []).map(entry => [entry.family, entry.count]),
  );

  return (
    <div className={root()}>
      {items.map((entry, index) => {
        const loads = loadsByFamily.get(entry.family);
        const count = denominators == null || loads == null
          ? String(entry.count)
          : t("trackerDiscovery.hitsOfLoads", { hits: entry.count, loads });

        return (
          <Fragment key={entry.family}>
            {compact && index > 0
              ? <Text size={2} color="faint" className={separator()}>·</Text>
              : null}
            <div className={item()}>
              <Text size={2} color={compact ? "faint" : "neutral"}>
                {familyLabel(entry.family, t)}
              </Text>
              <Text size={2} weight="medium" highContrast>
                {count}
              </Text>
            </div>
          </Fragment>
        );
      })}
    </div>
  );
}
