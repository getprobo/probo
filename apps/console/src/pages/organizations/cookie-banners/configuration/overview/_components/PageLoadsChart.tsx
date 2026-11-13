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

import { ChartBarIcon } from "@phosphor-icons/react";
import { ChartBandAxis } from "@probo/ui/src/v2/Chart/ChartBandAxis";
import { ChartFrame, type ChartMargin } from "@probo/ui/src/v2/Chart/ChartFrame";
import { ChartGridRows } from "@probo/ui/src/v2/Chart/ChartGridRows";
import { ChartLegend, ChartLegendItem } from "@probo/ui/src/v2/Chart/ChartLegend";
import { ChartStackedBars } from "@probo/ui/src/v2/Chart/ChartStackedBars";
import { ChartTooltip } from "@probo/ui/src/v2/Chart/ChartTooltip";
import { ChartValueAxis } from "@probo/ui/src/v2/Chart/ChartValueAxis";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { localPoint } from "@visx/event";
import { scaleBand, scaleLinear } from "@visx/scale";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { PageLoadsChart_cookieBanner$key } from "#/__generated__/core/PageLoadsChart_cookieBanner.graphql";

import { FamilyLogo } from "../../_components/FamilyLogo";
import { familyLabel } from "../../_lib/familyLabel";
import { cookieBannerOverview } from "../variants";

const families = [
  { key: "CHROME", fill: "fill-indigo-9", swatch: "bg-indigo-9" },
  { key: "SAFARI", fill: "fill-sky-9", swatch: "bg-sky-9" },
  { key: "EDGE", fill: "fill-green-9", swatch: "bg-green-9" },
  { key: "FIREFOX", fill: "fill-amber-9", swatch: "bg-amber-9" },
  { key: "OTHER", fill: "fill-sand-8", swatch: "bg-sand-8" },
] as const;

type Family = (typeof families)[number]["key"];

const familyKeys = new Set<string>(families.map(family => family.key));

const margin: ChartMargin = { top: 12, right: 4, bottom: 28, left: 44 };

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

interface DayRow extends Record<Family, number> {
  day: string;
  label: string;
  tooltipLabel: string;
}

function emptyCounts(): Record<Family, number> {
  return {
    CHROME: 0,
    SAFARI: 0,
    EDGE: 0,
    FIREFOX: 0,
    OTHER: 0,
  };
}

function familyOf(value: string): Family {
  if (familyKeys.has(value)) {
    return value as Family;
  }

  return "OTHER";
}

function dayTotal(row: DayRow, keys: readonly Family[]) {
  return keys.reduce((sum, key) => sum + row[key], 0);
}

export function PageLoadsChart({ cookieBannerKey }: PageLoadsChartProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(pageLoadsChartFragment, cookieBannerKey);
  const {
    plot,
    empty,
    emptyIcon,
    tooltipBody,
    tooltipRow,
    legendLabel,
  } = cookieBannerOverview();

  const dayFormat = new Intl.DateTimeFormat(i18n.language, {
    timeZone: "UTC",
    month: "short",
    day: "numeric",
  });
  const tooltipDayFormat = new Intl.DateTimeFormat(i18n.language, {
    timeZone: "UTC",
    weekday: "short",
    month: "short",
    day: "numeric",
  });
  const countFormat = new Intl.NumberFormat(i18n.language);

  const rows: DayRow[] = banner.dailyPageLoads.map((entry) => {
    const counts = emptyCounts();
    for (const load of entry.pageLoads) {
      counts[familyOf(load.family)] += load.count;
    }
    const day = new Date(entry.day);

    return {
      day: entry.day,
      label: dayFormat.format(day),
      tooltipLabel: tooltipDayFormat.format(day),
      ...counts,
    };
  });

  const totals = emptyCounts();
  for (const row of rows) {
    for (const family of families) {
      totals[family.key] += row[family.key];
    }
  }

  const keys = families
    .filter(family => totals[family.key] > 0)
    .map(family => family.key);
  const fill = Object.fromEntries(families.map(family => [family.key, family.fill]));
  const max = keys.length === 0
    ? 0
    : Math.max(...rows.map(row => dayTotal(row, keys)));

  if (max === 0) {
    return (
      <div className={empty()}>
        <ChartBarIcon size={32} className={emptyIcon()} />
        <Heading level={3} size={3} weight="medium" highContrast>
          {t("overview.chart.empty.title")}
        </Heading>
        <Text size={2} color="faint">
          {t("overview.chart.empty.description")}
        </Text>
      </div>
    );
  }

  const axisFormat = new Intl.NumberFormat(i18n.language, {
    notation: max >= 1000 ? "compact" : "standard",
    maximumFractionDigits: 1,
  });

  return (
    <div className={plot()}>
      <ChartTooltip<DayRow>
        content={row => (
          <div className={tooltipBody()}>
            <span>{row.tooltipLabel}</span>
            {keys.map(key => (
              <span key={key} className={tooltipRow()}>
                <span>{familyLabel(key, t)}</span>
                <span>{countFormat.format(row[key])}</span>
              </span>
            ))}
            <span className={tooltipRow()}>
              <span>{t("overview.chart.total")}</span>
              <span>{countFormat.format(dayTotal(row, keys))}</span>
            </span>
          </div>
        )}
      >
        {tooltip => (
          <ChartFrame
            height={240}
            margin={margin}
            aria-label={t("overview.chart.title")}
          >
            {({ width, height }) => {
              const xScale = scaleBand({
                domain: rows.map(row => row.label),
                range: [0, width],
                padding: rows.length > 14 ? 0.2 : 0.35,
              });
              const yScale = scaleLinear({
                domain: [0, max],
                range: [height, 0],
                nice: true,
              });

              return (
                <>
                  <ChartGridRows scale={yScale} width={width} />
                  <ChartStackedBars
                    data={rows}
                    keys={keys}
                    x={row => row.label}
                    xScale={xScale}
                    yScale={yScale}
                    height={height}
                    fill={fill}
                    onHover={(column, event) => {
                      if (column == null) {
                        tooltip.hide();
                        return;
                      }
                      const point = localPoint(event);
                      tooltip.show({
                        data: column.datum,
                        left: point?.x ?? 0,
                        top: point?.y ?? 0,
                      });
                    }}
                  />
                  <ChartBandAxis
                    scale={xScale}
                    top={height}
                    tickValues={rows.flatMap((row, index) => (
                      index % 5 === 0 || index === rows.length - 1 ? [row.label] : []
                    ))}
                  />
                  <ChartValueAxis
                    scale={yScale}
                    tickFormat={value => axisFormat.format(Number(value))}
                  />
                </>
              );
            }}
          </ChartFrame>
        )}
      </ChartTooltip>
      <ChartLegend>
        {families.filter(family => totals[family.key] > 0).map(family => (
          <ChartLegendItem
            key={family.key}
            swatchClassName={family.swatch}
            label={(
              <span className={legendLabel()}>
                <FamilyLogo
                  family={family.key}
                  size={16}
                  className={emptyIcon()}
                />
                <Text size={2}>{familyLabel(family.key, t)}</Text>
              </span>
            )}
            value={(
              <Text size={2} weight="medium" highContrast>
                {countFormat.format(totals[family.key])}
              </Text>
            )}
          />
        ))}
      </ChartLegend>
    </div>
  );
}
