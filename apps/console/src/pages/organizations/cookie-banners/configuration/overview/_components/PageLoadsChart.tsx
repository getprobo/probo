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

import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { PageLoadsChart_cookieBanner$key } from "#/__generated__/core/PageLoadsChart_cookieBanner.graphql";

import { FamilyLineChart } from "../../_components/FamilyLineChart";

const pageLoadsChartFragment = graphql`
  fragment PageLoadsChart_cookieBanner on CookieBanner {
    dailyPageLoads(days: 30) {
      day
      pageLoads {
        family
        count
      }
    }
  }
`;

interface PageLoadsChartProps {
  cookieBannerKey: PageLoadsChart_cookieBanner$key;
}

export function PageLoadsChart({ cookieBannerKey }: PageLoadsChartProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(pageLoadsChartFragment, cookieBannerKey);

  return (
    <FamilyLineChart
      days={banner.dailyPageLoads.map(entry => ({
        day: entry.day,
        counts: entry.pageLoads.map(load => ({
          family: load.family,
          count: load.count,
        })),
      }))}
      ariaLabel={t("overview.chart.title")}
      emptyTitle={t("overview.chart.empty.title")}
      emptyDescription={t("overview.chart.empty.description")}
      totalLabel={t("overview.chart.total")}
    />
  );
}
