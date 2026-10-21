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

import { useHotkeys } from "@probo/hotkeys";
import { useEffect, useSyncExternalStore } from "react";

interface FocusGroup {
  fields: HTMLInputElement[];
  target: HTMLInputElement | null;
  listeners: Set<() => void>;
  resizeObservers: Map<HTMLInputElement, ResizeObserver>;
}

const groups = new Map<string, FocusGroup>();
const focusHandlers = new Map<string, (event: KeyboardEvent) => void>();

let attributeObserver: MutationObserver | null = null;

function groupFor(key: string) {
  const existing = groups.get(key);
  if (existing != null) {
    return existing;
  }
  const group: FocusGroup = {
    fields: [],
    target: null,
    listeners: new Set(),
    resizeObservers: new Map(),
  };
  groups.set(key, group);
  return group;
}

function emit(group: FocusGroup) {
  for (const listener of group.listeners) {
    listener();
  }
}

function byDocumentOrder(left: HTMLInputElement, right: HTMLInputElement) {
  if (left === right) {
    return 0;
  }
  const position = left.compareDocumentPosition(right);
  if (position & Node.DOCUMENT_POSITION_FOLLOWING) {
    return -1;
  }
  if (position & Node.DOCUMENT_POSITION_PRECEDING) {
    return 1;
  }
  return 0;
}

function isFocusable(element: HTMLInputElement) {
  if (!element.isConnected || element.disabled) {
    return false;
  }
  if (element.closest("[inert]") != null) {
    return false;
  }
  return element.getClientRects().length > 0;
}

function firstFocusable(group: FocusGroup) {
  return group.fields.filter(isFocusable).sort(byDocumentOrder)[0] ?? null;
}

function refreshTarget(group: FocusGroup) {
  const next = firstFocusable(group);
  if (next === group.target) {
    return;
  }
  group.target = next;
  emit(group);
}

function syncAttributeObserver() {
  const active = [...groups.values()].some(group => group.fields.length > 0);
  if (typeof document === "undefined" || typeof MutationObserver === "undefined") {
    return;
  }
  if (active && attributeObserver == null) {
    attributeObserver = new MutationObserver(() => {
      for (const group of groups.values()) {
        refreshTarget(group);
      }
    });
    attributeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["disabled", "hidden", "inert"],
      subtree: true,
    });
    return;
  }
  if (!active && attributeObserver != null) {
    attributeObserver.disconnect();
    attributeObserver = null;
  }
}

function syncResizeObservers(group: FocusGroup) {
  for (const element of group.fields) {
    if (group.resizeObservers.has(element) || typeof ResizeObserver === "undefined") {
      continue;
    }
    const observer = new ResizeObserver(() => {
      refreshTarget(group);
    });
    observer.observe(element);
    group.resizeObservers.set(element, observer);
  }
  for (const [element, observer] of group.resizeObservers) {
    if (group.fields.includes(element)) {
      continue;
    }
    observer.disconnect();
    group.resizeObservers.delete(element);
  }
}

function publish(group: FocusGroup) {
  syncResizeObservers(group);
  refreshTarget(group);
  syncAttributeObserver();
  emit(group);
}

function focusHandler(key: string) {
  const existing = focusHandlers.get(key);
  if (existing != null) {
    return existing;
  }
  const handler = (event: KeyboardEvent) => {
    const group = groupFor(key);
    refreshTarget(group);
    const field = group.target;
    if (field == null) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    if (document.activeElement === field) {
      field.select();
      return;
    }
    field.focus();
  };
  focusHandlers.set(key, handler);
  return handler;
}

export function useFocusTarget(element: HTMLInputElement | null, key: string) {
  const group = groupFor(key);
  const isTarget = useSyncExternalStore(
    (listener) => {
      group.listeners.add(listener);
      return () => {
        group.listeners.delete(listener);
      };
    },
    () => element != null && group.target === element,
    () => false,
  );

  useEffect(() => {
    if (element == null) {
      return;
    }
    group.fields.push(element);
    publish(group);
    return () => {
      const index = group.fields.indexOf(element);
      if (index >= 0) {
        group.fields.splice(index, 1);
      }
      publish(group);
    };
  }, [element, group]);

  useHotkeys(`mod+${key}`, focusHandler(key), { capture: true });

  return isTarget;
}
