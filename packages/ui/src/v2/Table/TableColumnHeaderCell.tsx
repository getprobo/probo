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

import { CaretDownIcon, CaretUpDownIcon, CaretUpIcon } from "@phosphor-icons/react";
import type { ComponentProps } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { cellStyle } from "./cellStyle";
import { useTableContext } from "./context";
import { table } from "./variants";

export type TableColumnHeaderSort = "ascending" | "descending" | "none";

export type TableColumnHeaderCellProps
  = Omit<ComponentProps<"th">, "width">
    & Pick<VariantProps<typeof table>, "justify">
    & {
      width?: string;
      minWidth?: string;
      maxWidth?: string;
      // Controlled sort state. Maps to aria-sort. Ignored unless onSort is set.
      sort?: TableColumnHeaderSort;
      // Turns the header label into a sort button. The parent owns the next order.
      onSort?: () => void;
    };

// Column heading cell (Radix "Table.ColumnHeaderCell"). Renders a <th scope="col">.
// Pass onSort to make the label a controlled sort button; the kit does not toggle.
export function TableColumnHeaderCell(props: TableColumnHeaderCellProps) {
  const {
    justify,
    width,
    minWidth,
    maxWidth,
    className,
    style,
    sort,
    onSort,
    children,
    "aria-label": ariaLabel,
    ...rest
  } = props;
  const size = useTableContext();
  const sortState = onSort == null ? undefined : (sort ?? "none");
  const { cell, columnHeader, sortButton, sortIcon } = table({
    size,
    justify,
    sort: sortState,
  });
  const Icon = sortState === "ascending"
    ? CaretUpIcon
    : sortState === "descending"
      ? CaretDownIcon
      : CaretUpDownIcon;

  return (
    <th
      scope="col"
      aria-sort={sortState === "none" ? undefined : sortState}
      aria-label={onSort == null ? ariaLabel : undefined}
      className={columnHeader({ className: cell({ className }) })}
      style={cellStyle(width, minWidth, maxWidth, style)}
      {...rest}
    >
      {onSort == null
        ? children
        : (
            <button
              type="button"
              className={sortButton()}
              onClick={onSort}
              aria-label={ariaLabel}
            >
              {children}
              <Icon className={sortIcon()} />
            </button>
          )}
    </th>
  );
}
