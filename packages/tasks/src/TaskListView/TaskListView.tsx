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
import type { Virtualizer } from "@tanstack/react-virtual";
import { useVirtualizer } from "@tanstack/react-virtual";
import { type ReactNode, useEffect, useLayoutEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";

import { flattenTaskListItems } from "../flattenTaskListItems";
import { TASKS_NAMESPACE } from "../i18n";
import { taskListItemSize } from "../taskListSize";
import { isTaskState, type TaskState } from "../taskState";
import { type TaskListLoadNext, useTasksListFill } from "../useTasksListFill";
import { taskList, taskListView } from "../variants";

import { TaskListGroupHeader } from "./TaskListGroupHeader";
import { TaskListRowFrame } from "./TaskListRowFrame";
import {
  flatItemKey,
  listOffsetInScrollParent,
  nearestOverflowParent,
  pinnedTaskGroupHeader,
} from "./taskListViewLayout";

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
  const closedBeforeSearchRef = useRef<Set<TaskState> | null>(null);
  const openedSearchKeyRef = useRef<string | null>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const [scrollElement, setScrollElement] = useState<HTMLElement | null>(null);
  const [listOffset, setListOffset] = useState(0);
  const [, startTransition] = useTransition();
  const visibleGroups = groups.filter(
    group => selectedState == null || group.state === selectedState,
  );
  const resultStatesKey = visibleGroups
    .map(group => `${group.state}:${group.totalCount > 0 ? 1 : 0}`)
    .join("|");
  const isFilling = visibleGroups.some(group => group.isLoadingNext);

  useEffect(() => {
    if (query === "") {
      openedSearchKeyRef.current = null;
      if (closedBeforeSearchRef.current != null) {
        setClosedStates(closedBeforeSearchRef.current);
        closedBeforeSearchRef.current = null;
      }
      return;
    }

    const key = `${query}:${resultStatesKey}`;
    if (openedSearchKeyRef.current === key) {
      return;
    }
    openedSearchKeyRef.current = key;

    const openStates = resultStatesKey.split("|").flatMap((part) => {
      const separator = part.lastIndexOf(":");
      if (separator === -1 || part.slice(separator + 1) !== "1") {
        return [];
      }
      const state = part.slice(0, separator);
      return isTaskState(state) ? [state] : [];
    });

    setClosedStates((current) => {
      if (closedBeforeSearchRef.current == null) {
        closedBeforeSearchRef.current = new Set(current);
      }
      const next = new Set(current);
      let changed = false;
      for (const state of openStates) {
        if (next.delete(state)) {
          changed = true;
        }
      }
      return changed ? next : current;
    });
  }, [query, resultStatesKey]);
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
  const headerIndexes = useMemo(() => {
    const indexes: number[] = [];
    for (let index = 0; index < flatItems.length; index += 1) {
      if (flatItems[index].type === "header") {
        indexes.push(index);
      }
    }
    return indexes;
  }, [flatItems]);
  const flatItemsRef = useRef(flatItems);
  const headerIndexesRef = useRef(headerIndexes);
  const listOffsetRef = useRef(listOffset);
  const pinShiftRef = useRef<HTMLDivElement>(null);
  const pinnedShiftRef = useRef(0);
  flatItemsRef.current = flatItems;
  headerIndexesRef.current = headerIndexes;
  listOffsetRef.current = listOffset;
  const [pinnedIndex, setPinnedIndex] = useState<number | null>(null);
  // eslint-disable-next-line react-hooks/incompatible-library
  const virtualizer = useVirtualizer({
    count: flatItems.length,
    getScrollElement: () => scrollElement,
    estimateSize: index => taskListItemSize(flatItems[index]),
    overscan: 24,
    scrollMargin: listOffset,
    directDomUpdates: true,
    getItemKey: index => flatItemKey(flatItems[index], index),
    onChange(instance) {
      publishPinnedHeader(instance);
    },
  });

  function publishPinnedHeader(instance: Virtualizer<HTMLElement, Element>) {
    const offset = instance.scrollOffset;
    if (offset == null) {
      return;
    }
    const scrolledIntoList = offset - listOffsetRef.current;
    if (scrolledIntoList <= 0) {
      pinnedShiftRef.current = 0;
      if (pinShiftRef.current != null) {
        pinShiftRef.current.style.transform = "";
      }
      setPinnedIndex(current => current === null ? current : null);
      return;
    }
    const next = pinnedTaskGroupHeader(
      flatItemsRef.current,
      headerIndexesRef.current,
      instance.measurementsCache,
      scrolledIntoList,
      instance.options.scrollMargin,
    );
    const shift = next?.shift ?? 0;
    pinnedShiftRef.current = shift;
    const node = pinShiftRef.current;
    if (node != null) {
      node.style.transform = shift === 0 ? "" : `translateY(${shift}px)`;
    }
    const nextIndex = next?.index ?? null;
    setPinnedIndex(current => current === nextIndex ? current : nextIndex);
  }
  const virtualItems = virtualizer.getVirtualItems();
  const pinnedItem = pinnedIndex == null ? null : flatItems[pinnedIndex];
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

  const hasRows = flatItems.length > 0;
  useEffect(() => {
    setScrollElement(nearestOverflowParent(listRef.current));
  }, [hasRows]);

  useLayoutEffect(() => {
    const list = listRef.current;
    if (list == null || scrollElement == null) {
      return;
    }

    const listElement = list;
    const scroll = scrollElement;
    function measure() {
      const next = listOffsetInScrollParent(scroll, listElement);
      setListOffset(current => (Math.abs(current - next) < 0.5 ? current : next));
    }

    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(scrollElement);
    for (const child of scrollElement.children) {
      observer.observe(child);
    }

    return () => {
      observer.disconnect();
    };
  }, [flatItems, scrollElement]);

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
        {pinnedItem?.type === "header" && pinnedIndex != null && (
          <div className="sticky top-0 z-2 h-0">
            <div
              ref={pinShiftRef}
              style={pinnedShiftRef.current === 0
                ? undefined
                : { transform: `translateY(${pinnedShiftRef.current}px)` }}
            >
              <TaskListGroupHeader
                state={pinnedItem.state}
                count={pinnedItem.count}
                open={!closedStates.has(pinnedItem.state)}
                offset={false}
                pinned
                onOpenChange={open => onGroupOpenChange(pinnedItem.state, open)}
              />
            </div>
          </div>
        )}
        <div
          ref={virtualizer.containerRef}
          className="relative w-full"
        >
          {virtualItems.map((virtualItem) => {
            const item = flatItems[virtualItem.index];
            const pinnedAway = pinnedIndex != null && virtualItem.index === pinnedIndex;
            const leading = item.type !== "header"
              && (item.position === "first" || item.position === "only");
            const trailing = item.type !== "header"
              && (item.position === "last" || item.position === "only");
            return (
              <div
                key={virtualItem.key}
                data-index={virtualItem.index}
                ref={virtualizer.measureElement}
                className={`absolute top-0 left-0 w-full overflow-hidden${leading ? " pt-2" : ""}${trailing ? " pb-2" : ""}${pinnedAway ? " invisible" : ""}`}
                aria-hidden={pinnedAway || undefined}
                style={{ height: virtualItem.size }}
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
