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

import type { ComponentProps } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { TextSkeleton } from "../typography/TextSkeleton";

import { table } from "./variants";

export type TableSkeletonProps = Omit<ComponentProps<"div">, "children">
  & Pick<VariantProps<typeof table>, "size" | "variant" | "layout">
  & {
    // Number of pulse body rows to render inside the table surface.
    count?: number;
    // Number of pulse columns per row, including the header.
    columns?: number;
  };

// Loading placeholder paired with Table: the same frame filled with pulse
// header and body cells. Imports only variants + TextSkeleton so it stays out
// of any interactive Table consumer graph.
export function TableSkeleton(props: TableSkeletonProps) {
  const {
    size = 2,
    variant = "ghost",
    layout = "auto",
    count = 3,
    columns = 4,
    className,
    ...rest
  } = props;
  const rows = Array.from({ length: count }, (_, index) => index);
  const cells = Array.from({ length: columns }, (_, index) => index);
  const { root, table: tableSlot, header, body, row, cell, columnHeader } = table({
    size,
    variant,
    layout,
  });

  return (
    <div className={root({ className })} aria-hidden {...rest}>
      <table className={tableSlot()}>
        <thead className={header()}>
          <tr className={row()}>
            {cells.map(column => (
              <th key={column} className={columnHeader({ className: cell() })}>
                <TextSkeleton size={size === 3 ? 3 : 2} className="w-16" />
              </th>
            ))}
          </tr>
        </thead>
        <tbody className={body()}>
          {rows.map(rowIndex => (
            <tr key={rowIndex} className={row()}>
              {cells.map(column => (
                <td key={column} className={cell()}>
                  <TextSkeleton size={size === 3 ? 3 : 2} className="w-24" />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
