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

import type { ScaleInput } from "@visx/scale";
import { BarStack, getBandwidth, type PositionScale, type StackKey } from "@visx/shape";
import type { MouseEvent } from "react";

import { chart } from "./variants";

export interface ChartColumnHover<Datum> {
  index: number;
  datum: Datum;
  x: number;
  width: number;
}

export interface ChartStackedBarsProps<
  Datum extends object,
  Key extends StackKey,
  XScale extends PositionScale,
  YScale extends PositionScale,
> {
  data: Datum[];
  keys: Key[];
  x: (datum: Datum) => ScaleInput<XScale>;
  xScale: XScale;
  yScale: YScale;
  height: number;
  fill: Partial<Record<Key, string>>;
  onHover?: (
    column: ChartColumnHover<Datum> | null,
    event: MouseEvent<SVGRectElement>,
  ) => void;
}

export function ChartStackedBars<
  Datum extends object,
  Key extends StackKey,
  XScale extends PositionScale,
  YScale extends PositionScale,
>({
  data,
  keys,
  x,
  xScale,
  yScale,
  height,
  fill,
  onHover,
}: ChartStackedBarsProps<Datum, Key, XScale, YScale>) {
  const { bar: barClass } = chart();

  return (
    <g>
      <BarStack
        data={data}
        keys={keys}
        x={x}
        xScale={xScale}
        yScale={yScale}
        color={() => "currentColor"}
      >
        {barStacks => barStacks.map(barStack => barStack.bars.map(bar => (
          <rect
            key={`bar-${barStack.index}-${bar.index}`}
            x={bar.x}
            y={bar.y}
            width={bar.width}
            height={Math.max(0, bar.height)}
            className={barClass({ className: fill[bar.key] })}
          />
        )))}
      </BarStack>
      {data.map((datum, index) => {
        const columnX = Number(xScale(x(datum)) ?? 0);
        const columnWidth = getBandwidth(xScale);

        return (
          <rect
            key={`column-${index}`}
            x={columnX}
            y={0}
            width={columnWidth}
            height={height}
            fill="transparent"
            onMouseMove={event => onHover?.(
              { index, datum, x: columnX, width: columnWidth },
              event,
            )}
            onMouseLeave={event => onHover?.(null, event)}
          />
        );
      })}
    </g>
  );
}
