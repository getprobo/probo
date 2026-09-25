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

import { TableProvider } from "./context";
import { table } from "./variants";

export type TableProps = ComponentProps<"div"> & Pick<VariantProps<typeof table>, "size" | "variant" | "layout">;

// Scroll wrapper around a semantic <table> (Radix "Table.Root"). Size and
// variant live here so cells inherit padding; pair with TableHeader / TableBody
// / TableRow / TableCell. Use TableSkeleton while loading.
export function Table(props: TableProps) {
  const { size = 2, variant = "ghost", layout = "auto", className, children, ...rest } = props;
  const { root, table: tableSlot } = table({ size, variant, layout });

  return (
    <div className={root({ className })} {...rest}>
      <TableProvider value={size}>
        <table className={tableSlot()}>{children}</table>
      </TableProvider>
    </div>
  );
}
