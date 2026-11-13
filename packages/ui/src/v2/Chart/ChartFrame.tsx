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

import { Group } from "@visx/group";
import { useParentSize } from "@visx/responsive";
import type { ComponentProps, ReactNode } from "react";

import { chart } from "./variants";

export interface ChartMargin {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

export interface ChartFrameSize {
  width: number;
  height: number;
}

export type ChartFrameProps = Omit<ComponentProps<"svg">, "children" | "width" | "height"> & {
  height: number;
  margin: ChartMargin;
  children: (size: ChartFrameSize) => ReactNode;
};

export function ChartFrame({
  height,
  margin,
  children,
  ...svgProps
}: ChartFrameProps) {
  const { parentRef, width } = useParentSize({ debounceTime: 0 });
  const { frame, svg } = chart();
  const innerWidth = Math.max(0, width - margin.left - margin.right);
  const innerHeight = Math.max(0, height - margin.top - margin.bottom);

  return (
    <div ref={parentRef} className={frame()}>
      {width > 0
        ? (
            <svg
              width={width}
              height={height}
              className={svg()}
              {...svgProps}
            >
              <Group left={margin.left} top={margin.top}>
                {children({ width: innerWidth, height: innerHeight })}
              </Group>
            </svg>
          )
        : null}
    </div>
  );
}
