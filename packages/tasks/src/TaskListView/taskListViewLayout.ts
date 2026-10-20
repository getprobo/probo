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

import type { TaskListFlatItem } from "../flattenTaskListItems";
import {
  taskListGroupHeaderOffset,
  taskListGroupHeaderSize,
} from "../taskListSize";

export function nearestOverflowParent(element: HTMLElement | null): HTMLElement | null {
  let current = element?.parentElement ?? null;
  while (current != null) {
    const overflowY = getComputedStyle(current).overflowY;
    if (overflowY === "auto" || overflowY === "scroll") {
      return current;
    }
    current = current.parentElement;
  }
  return null;
}

const pinnedHeaderHeight = taskListGroupHeaderSize;

interface PinnedTaskGroupHeader {
  index: number;
  shift: number;
}

export function pinnedTaskGroupHeader<T>(
  items: readonly TaskListFlatItem<T>[],
  headerIndexes: readonly number[],
  measurements: readonly { start: number }[],
  scrolledIntoList: number,
  scrollMargin: number,
): PinnedTaskGroupHeader | null {
  if (measurements.length !== items.length) {
    return null;
  }

  let activeIndex = -1;
  let nextStart: number | null = null;
  for (const index of headerIndexes) {
    const item = items[index];
    if (item.type !== "header") {
      continue;
    }

    const start = measurements[index].start - scrollMargin;
    const contentStart = start + (item.offset ? taskListGroupHeaderOffset : 0);
    if (contentStart > scrolledIntoList) {
      if (activeIndex >= 0) {
        nextStart = start;
      }
      break;
    }
    activeIndex = index;
  }

  if (activeIndex < 0) {
    return null;
  }

  const shift = nextStart == null
    ? 0
    : Math.min(0, nextStart - scrolledIntoList - pinnedHeaderHeight);

  return { index: activeIndex, shift };
}

export function listOffsetInScrollParent(scroll: HTMLElement, list: HTMLElement): number {
  return list.getBoundingClientRect().top - scroll.getBoundingClientRect().top + scroll.scrollTop;
}

export function flatItemKey<T extends { id: string }>(
  item: TaskListFlatItem<T>,
  index: number,
): string {
  if (item.type === "header") {
    return `header:${item.state}`;
  }
  if (item.type === "row") {
    return item.item.id;
  }
  return `placeholder:${item.state}:${index}`;
}
