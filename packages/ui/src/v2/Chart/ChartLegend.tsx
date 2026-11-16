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

import type { ComponentProps, ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { chart } from "./variants";

type ChartLegendProps = ComponentProps<"ul"> & Pick<VariantProps<typeof chart>, "orientation">;

export function ChartLegend({ orientation, className, ...rest }: ChartLegendProps) {
  const { legend } = chart({ orientation });

  return <ul className={legend({ className })} {...rest} />;
}

export interface ChartLegendItemProps extends Pick<ComponentProps<"li">, "onMouseEnter" | "onMouseLeave" | "className"> {
  swatchClassName?: string;
  swatchSize?: "sm" | "md";
  itemAlign?: "center" | "stretch";
  // Replaces the color swatch. Used when the mark is a logo on a circle.
  mark?: ReactNode;
  label: ReactNode;
  value?: ReactNode;
}

export function ChartLegendItem({
  swatchClassName,
  swatchSize,
  itemAlign,
  mark,
  label,
  value,
  className,
  onMouseEnter,
  onMouseLeave,
}: ChartLegendItemProps) {
  const { legendItem, swatch } = chart({ swatchSize, itemAlign });

  return (
    <li
      className={legendItem({ className })}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
    >
      {mark ?? <span className={swatch({ className: swatchClassName })} />}
      <span>{label}</span>
      {value != null && <span>{value}</span>}
    </li>
  );
}
