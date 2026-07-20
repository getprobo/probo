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

import { TextSkeleton } from "../typography/TextSkeleton";

import { list } from "./variants";

export type ListItemSkeletonProps = Omit<ComponentProps<"li">, "children">;

// Loading placeholder paired with ListItem: same row shell (card background +
// dividers) with pulse bars for the primary column — not a filled sand-2 slab.
export function ListItemSkeleton(props: ListItemSkeletonProps) {
  const { className, ...rest } = props;
  const { item, content } = list();

  return (
    <li className={item({ className })} aria-hidden {...rest}>
      <div className={content()}>
        <TextSkeleton size={2} className="w-40" />
        <TextSkeleton size={1} className="w-24" />
      </div>
    </li>
  );
}
