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

import { ChartLineIcon } from "@phosphor-icons/react";
import { ChartBandAxis } from "@probo/ui/src/v2/Chart/ChartBandAxis";
import { ChartFrame, type ChartMargin } from "@probo/ui/src/v2/Chart/ChartFrame";
import { ChartGridRows } from "@probo/ui/src/v2/Chart/ChartGridRows";
import { ChartLegend, ChartLegendItem } from "@probo/ui/src/v2/Chart/ChartLegend";
import { ChartLines } from "@probo/ui/src/v2/Chart/ChartLines";
import { ChartTooltip } from "@probo/ui/src/v2/Chart/ChartTooltip";
import { ChartValueAxis } from "@probo/ui/src/v2/Chart/ChartValueAxis";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { localPoint } from "@visx/event";
import { scaleLog, scalePoint } from "@visx/scale";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { familyLabel } from "../_lib/familyLabel";
import { cookieBannerOverview } from "../overview/variants";

import { FamilyLogo } from "./FamilyLogo";

const families = [
  { key: "CHROME", stroke: "stroke-indigo-9", point: "fill-indigo-9", swatch: "bg-indigo-9" },
  { key: "SAFARI", stroke: "stroke-sky-9", point: "fill-sky-9", swatch: "bg-sky-9" },
  { key: "EDGE", stroke: "stroke-green-9", point: "fill-green-9", swatch: "bg-green-9" },
  { key: "FIREFOX", stroke: "stroke-amber-9", point: "fill-amber-9", swatch: "bg-amber-9" },
  { key: "OTHER", stroke: "stroke-sand-8", point: "fill-sand-8", swatch: "bg-sand-8" },
] as const;

type Family = (typeof families)[number]["key"];

const familyKeys = new Set<string>(families.map(family => family.key));

const margin: ChartMargin = { top: 12, right: 4, bottom: 16, left: 44 };

export interface FamilyLineChartDay {
  day: string;
  counts: readonly { family: string; count: number }[];
}

interface FamilyLineChartProps {
  days: readonly FamilyLineChartDay[];
  ariaLabel: string;
  emptyTitle: string;
  emptyDescription: string;
  totalLabel: string;
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

function logTop(max: number) {
  const top = 10 ** Math.ceil(Math.log10(Math.max(max, 1)));

  return top <= 1 ? 10 : top;
}

function powersOfTen(top: number) {
  const ticks: number[] = [];
  for (let tick = 1; tick <= top; tick *= 10) {
    ticks.push(tick);
  }

  return ticks;
}

export function FamilyLineChart({
  days,
  ariaLabel,
  emptyTitle,
  emptyDescription,
  totalLabel,
}: FamilyLineChartProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const {
    plot,
    empty,
    emptyIcon,
    tooltipBody,
    tooltipRow,
    legendCopy,
    legendMark,
    legendSwatch,
    legendLogo,
    legendTotal,
    legendItem,
    legendColumn,
    chartPane,
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
  const [emphasized, setEmphasized] = useState<Family | null>(null);

  const rows: DayRow[] = days.map((entry) => {
    const counts = emptyCounts();
    for (const count of entry.counts) {
      counts[familyOf(count.family)] += count.count;
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
  const stroke = Object.fromEntries(families.map(family => [family.key, family.stroke]));
  const point = Object.fromEntries(families.map(family => [family.key, family.point]));
  const max = keys.length === 0
    ? 0
    : Math.max(...rows.flatMap(row => keys.map(key => row[key])));

  if (max === 0) {
    return (
      <div className={empty()}>
        <ChartLineIcon size={32} className={emptyIcon()} />
        <Heading level={3} size={3} weight="medium" highContrast>
          {emptyTitle}
        </Heading>
        <Text size={2} color="faint">
          {emptyDescription}
        </Text>
      </div>
    );
  }

  const top = logTop(max);
  const axisFormat = new Intl.NumberFormat(i18n.language, {
    notation: top >= 1000 ? "compact" : "standard",
    maximumFractionDigits: 1,
  });

  return (
    <div className={plot()}>
      <ChartLegend orientation="vertical" className={legendColumn()}>
        {families.filter(family => totals[family.key] > 0).map(family => (
          <ChartLegendItem
            key={family.key}
            className={legendItem()}
            onMouseEnter={() => setEmphasized(family.key)}
            onMouseLeave={() => setEmphasized(null)}
            itemAlign="center"
            mark={(
              <span className={legendMark()}>
                <span className={legendSwatch({ className: family.swatch })} />
                <FamilyLogo family={family.key} size={16} className={legendLogo()} />
              </span>
            )}
            label={(
              <span className={legendCopy()}>
                <Text size={1}>{familyLabel(family.key, t)}</Text>
                <Text size={1} weight="bold" highContrast className={legendTotal()}>
                  {countFormat.format(totals[family.key])}
                </Text>
              </span>
            )}
          />
        ))}
      </ChartLegend>
      <div className={chartPane()}>
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
                <span>{totalLabel}</span>
                <span>{countFormat.format(dayTotal(row, keys))}</span>
              </span>
            </div>
          )}
        >
          {tooltip => (
            <ChartFrame
              height={240}
              margin={margin}
              aria-label={ariaLabel}
            >
              {({ width, height }) => {
                const xScale = scalePoint({
                  domain: rows.map(row => row.label),
                  range: [0, width],
                  padding: 0.5,
                });
                const yScale = scaleLog({
                  domain: [1, top],
                  range: [height, 0],
                  base: 10,
                  clamp: true,
                });

                return (
                  <>
                    <ChartGridRows
                      scale={yScale}
                      width={width}
                      tickValues={powersOfTen(top)}
                    />
                    <ChartLines
                      data={rows}
                      keys={keys}
                      x={row => row.label}
                      y={(row, key) => row[key]}
                      xScale={xScale}
                      yScale={yScale}
                      height={height}
                      stroke={stroke}
                      point={point}
                      emphasized={emphasized}
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
                      tickLength={4}
                      tickValues={rows.map(row => row.label)}
                      tickFormat={(value, index, values) => (
                        index % 5 === 0 || index === values.length - 1 ? String(value) : ""
                      )}
                    />
                    <ChartValueAxis
                      scale={yScale}
                      tickValues={powersOfTen(top)}
                      tickFormat={value => axisFormat.format(Number(value))}
                    />
                  </>
                );
              }}
            </ChartFrame>
          )}
        </ChartTooltip>
      </div>
    </div>
  );
}
