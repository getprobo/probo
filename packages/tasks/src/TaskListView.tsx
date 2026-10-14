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

import { TooltipProvider } from "@probo/ui/src/v2/Tooltip/TooltipProvider";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useVirtualizer } from "@tanstack/react-virtual";
import { type ReactNode, useEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";

import { flattenTaskListItems, type TaskListFlatItem } from "./flattenTaskListItems";
import { TASKS_NAMESPACE } from "./i18n";
import { TaskListGroupHeader } from "./TaskListGroupHeader";
import { TaskListRowFrame } from "./TaskListRowFrame";
import { taskListItemSize } from "./taskListSize";
import type { TaskState } from "./taskState";
import { type TaskListLoadNext, useTasksListFill } from "./useTasksListFill";
import { taskList, taskListView } from "./variants";

export interface TaskListViewItem {
  id: string;
}

export interface TaskListViewGroup<T extends TaskListViewItem> {
  state: TaskState;
  totalCount: number;
  loadedItems: readonly T[];
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: TaskListLoadNext;
}

export interface TaskListViewProps<T extends TaskListViewItem> {
  groups: readonly TaskListViewGroup<T>[];
  selectedState: TaskState | null;
  query: string;
  filtered?: boolean;
  busy?: boolean;
  renderRow: (item: T, sectionState: TaskState) => ReactNode;
}

function nearestOverflowParent(element: HTMLElement | null): HTMLElement | null {
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

function flatItemKey<T extends TaskListViewItem>(
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

export function TaskListView<T extends TaskListViewItem>({
  groups,
  selectedState,
  query,
  filtered = false,
  busy = false,
  renderRow,
}: TaskListViewProps<T>) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [closedStates, setClosedStates] = useState(
    () => new Set<TaskState>(["CANCELED", "DUPLICATE"]),
  );
  const listRef = useRef<HTMLDivElement>(null);
  const [scrollElement, setScrollElement] = useState<HTMLElement | null>(null);
  const [, startTransition] = useTransition();
  const visibleGroups = groups.filter(
    group => selectedState == null || group.state === selectedState,
  );
  const isFilling = visibleGroups.some(group => group.isLoadingNext);
  const slots = taskList();
  const flatItems = useMemo(
    () => flattenTaskListItems(
      visibleGroups.map(group => ({
        state: group.state,
        totalCount: group.totalCount,
        loadedItems: group.loadedItems,
        open: !closedStates.has(group.state),
      })),
    ),
    [closedStates, visibleGroups],
  );
  const virtualizer = useVirtualizer({
    count: flatItems.length,
    getScrollElement: () => scrollElement,
    estimateSize: index => taskListItemSize(flatItems[index]),
    overscan: 8,
    getItemKey: index => flatItemKey(flatItems[index], index),
  });
  const virtualItems = virtualizer.getVirtualItems();
  const fillRange = virtualItems.length === 0
    ? null
    : {
        startIndex: virtualItems[0].index,
        endIndex: virtualItems[virtualItems.length - 1].index,
      };
  useTasksListFill({
    items: flatItems,
    groups: visibleGroups.map(group => ({
      state: group.state,
      totalCount: group.totalCount,
      loadedCount: group.loadedItems.length,
      hasNext: group.hasNext,
      loadNext: group.loadNext,
    })),
    range: fillRange,
    closedStates,
    selectedState,
    startTransition,
  });

  useEffect(() => {
    setScrollElement(nearestOverflowParent(listRef.current));
  }, []);

  function onGroupOpenChange(state: TaskState, open: boolean) {
    setClosedStates((current) => {
      const next = new Set(current);
      if (open) {
        next.delete(state);
      } else {
        next.add(state);
      }
      return next;
    });
  }

  if (flatItems.length === 0) {
    return (
      <div
        className={taskListView({ busy: busy || isFilling })}
        aria-busy={busy || isFilling || undefined}
      >
        <Text size={2} color="faint" align="center" className={slots.empty()}>
          {filtered || query !== "" || selectedState != null ? t("noResults") : t("empty")}
        </Text>
      </div>
    );
  }

  return (
    <div
      ref={listRef}
      className={taskListView({ busy: busy || isFilling })}
      aria-busy={busy || isFilling || undefined}
    >
      <TooltipProvider>
        <div
          className="relative w-full"
          style={{ height: virtualizer.getTotalSize() }}
        >
          {virtualizer.getVirtualItems().map((virtualItem) => {
            const item = flatItems[virtualItem.index];
            return (
              <div
                key={virtualItem.key}
                data-index={virtualItem.index}
                ref={virtualizer.measureElement}
                className="absolute top-0 left-0 w-full"
                style={{ transform: `translateY(${virtualItem.start}px)` }}
              >
                {item.type === "header"
                  ? (
                      <TaskListGroupHeader
                        state={item.state}
                        count={item.count}
                        open={!closedStates.has(item.state)}
                        offset={item.offset}
                        onOpenChange={open => onGroupOpenChange(item.state, open)}
                      />
                    )
                  : (
                      <TaskListRowFrame
                        position={item.position}
                        placeholder={item.type === "placeholder"}
                      >
                        {item.type === "row" && renderRow(item.item, item.state)}
                      </TaskListRowFrame>
                    )}
              </div>
            );
          })}
        </div>
      </TooltipProvider>
    </div>
  );
}
