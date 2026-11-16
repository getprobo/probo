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
import { getBandwidth, LinePath, type PositionScale } from "@visx/shape";
import { type MouseEvent, useState } from "react";

import type { ChartColumnHover } from "./ChartStackedBars";
import { chart } from "./variants";

type SteppedScale = PositionScale & { step?: () => number };

export interface ChartLinesProps<
  Datum extends object,
  Key extends string,
  XScale extends PositionScale,
  YScale extends PositionScale,
> {
  data: Datum[];
  keys: Key[];
  x: (datum: Datum) => ScaleInput<XScale>;
  y: (datum: Datum, key: Key) => number;
  xScale: XScale;
  yScale: YScale;
  height: number;
  stroke: Partial<Record<Key, string>>;
  point: Partial<Record<Key, string>>;
  onHover?: (
    column: ChartColumnHover<Datum> | null,
    event: MouseEvent<SVGRectElement>,
  ) => void;
}

const pointRadius = 2.5;
const hoveredPointRadius = 4;

function position<Scale extends PositionScale>(scale: Scale, value: ScaleInput<Scale>) {
  const raw = scale(value);

  return typeof raw === "number" ? raw : 0;
}

function columnSpan<XScale extends SteppedScale>(scale: XScale, value: ScaleInput<XScale>) {
  const at = position(scale, value);
  const bandwidth = getBandwidth(scale);
  if (bandwidth > 0) {
    return { x: at, width: bandwidth };
  }

  const step = scale.step?.() ?? 0;

  return { x: at - step / 2, width: step };
}

export function ChartLines<
  Datum extends object,
  Key extends string,
  XScale extends PositionScale,
  YScale extends PositionScale,
>({
  data,
  keys,
  x,
  y,
  xScale,
  yScale,
  height,
  stroke,
  point,
  onHover,
}: ChartLinesProps<Datum, Key, XScale, YScale>) {
  const { series, point: pointClass } = chart();
  const [hovered, setHovered] = useState<number | null>(null);

  return (
    <g>
      {keys.map(key => (
        <LinePath
          key={key}
          data={data}
          defined={datum => y(datum, key) > 0}
          x={datum => position(xScale, x(datum))}
          y={datum => position(yScale, y(datum, key))}
          stroke="currentColor"
          strokeWidth={2}
          className={series({ className: stroke[key] })}
        />
      ))}
      {data.map((datum, index) => keys.map((key) => {
        const value = y(datum, key);
        if (value <= 0) {
          return null;
        }

        return (
          <circle
            key={`${key}-${index}`}
            cx={position(xScale, x(datum))}
            cy={position(yScale, value)}
            r={index === hovered ? hoveredPointRadius : pointRadius}
            className={pointClass({ className: point[key] })}
          />
        );
      }))}
      {data.map((datum, index) => {
        const column = columnSpan(xScale, x(datum));

        return (
          <rect
            key={`column-${index}`}
            x={column.x}
            y={0}
            width={column.width}
            height={height}
            fill="transparent"
            onMouseMove={(event) => {
              setHovered(index);
              onHover?.(
                { index, datum, x: column.x, width: column.width },
                event,
              );
            }}
            onMouseLeave={(event) => {
              setHovered(null);
              onHover?.(null, event);
            }}
          />
        );
      })}
    </g>
  );
}
