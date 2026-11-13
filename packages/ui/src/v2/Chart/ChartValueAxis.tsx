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

import { AxisLeft, type AxisScale, type TickFormatter } from "@visx/axis";

import { chart } from "./variants";

export interface ChartValueAxisProps<Scale extends AxisScale> {
  scale: Scale;
  numTicks?: number;
  tickFormat?: TickFormatter<number>;
}

export function ChartValueAxis<Scale extends AxisScale>({
  scale,
  numTicks = 4,
  tickFormat,
}: ChartValueAxisProps<Scale>) {
  const { tick, line } = chart();

  return (
    <AxisLeft
      scale={scale}
      numTicks={numTicks}
      tickFormat={tickFormat}
      tickLength={4}
      stroke="currentColor"
      tickStroke="currentColor"
      axisLineClassName={line()}
      tickLineProps={{ className: line() }}
      tickClassName={tick()}
      tickLabelProps={{ className: tick(), fill: "currentColor" }}
    />
  );
}
