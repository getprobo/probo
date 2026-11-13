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

import type { Meta, StoryObj } from "@storybook/react";
import { localPoint } from "@visx/event";
import { scaleBand, scaleLinear } from "@visx/scale";

import { ChartBandAxis } from "./ChartBandAxis";
import { ChartFrame, type ChartMargin } from "./ChartFrame";
import { ChartGridRows } from "./ChartGridRows";
import { ChartLegend, ChartLegendItem } from "./ChartLegend";
import { ChartSkeleton } from "./ChartSkeleton";
import { ChartStackedBars } from "./ChartStackedBars";
import { ChartTooltip } from "./ChartTooltip";
import { ChartValueAxis } from "./ChartValueAxis";

const margin: ChartMargin = { top: 8, right: 8, bottom: 28, left: 36 };
const keys = ["chrome", "safari", "firefox"] as const;

type StoryDatum = {
  x: string;
  chrome: number;
  safari: number;
  firefox: number;
};

const data: StoryDatum[] = [
  { x: "Mon", chrome: 18, safari: 6, firefox: 3 },
  { x: "Tue", chrome: 22, safari: 8, firefox: 4 },
  { x: "Wed", chrome: 16, safari: 5, firefox: 7 },
  { x: "Thu", chrome: 28, safari: 9, firefox: 4 },
  { x: "Fri", chrome: 24, safari: 7, firefox: 6 },
  { x: "Sat", chrome: 12, safari: 4, firefox: 2 },
  { x: "Sun", chrome: 10, safari: 3, firefox: 2 },
];

const fill = {
  chrome: "fill-indigo-9",
  safari: "fill-sky-9",
  firefox: "fill-amber-9",
};

const swatch = {
  chrome: "bg-indigo-9",
  safari: "bg-sky-9",
  firefox: "bg-amber-9",
};

function total(datum: StoryDatum) {
  return datum.chrome + datum.safari + datum.firefox;
}

function StackedBarsChart() {
  const max = Math.max(...data.map(total));

  return (
    <div className="flex w-[36rem] flex-col gap-4">
      <ChartTooltip<StoryDatum>
        content={datum => (
          <div className="flex flex-col gap-0.5">
            <span>{datum.x}</span>
            <span>{total(datum)}</span>
          </div>
        )}
      >
        {tooltip => (
          <ChartFrame height={240} margin={margin} aria-label="Page loads by day">
            {({ width, height }) => {
              const xScale = scaleBand({
                domain: data.map(datum => datum.x),
                range: [0, width],
                padding: 0.35,
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
                    data={data}
                    keys={[...keys]}
                    x={datum => datum.x}
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
                  <ChartBandAxis scale={xScale} top={height} />
                  <ChartValueAxis scale={yScale} />
                </>
              );
            }}
          </ChartFrame>
        )}
      </ChartTooltip>
      <ChartLegend>
        {keys.map(key => (
          <ChartLegendItem
            key={key}
            swatchClassName={swatch[key]}
            label={key}
            value={data.reduce((sum, datum) => sum + datum[key], 0)}
          />
        ))}
      </ChartLegend>
    </div>
  );
}

export default {
  title: "v2/Chart",
  component: ChartFrame,
} satisfies Meta<typeof ChartFrame>;

type Story = StoryObj<typeof ChartFrame>;

export const StackedBars: Story = {
  render: () => <StackedBarsChart />,
};

export const Skeleton: Story = {
  render: () => (
    <div className="w-[36rem]">
      <ChartSkeleton />
    </div>
  ),
};
