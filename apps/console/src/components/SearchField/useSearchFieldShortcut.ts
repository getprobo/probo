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

import { useEffect, useSyncExternalStore } from "react";

const fields: HTMLInputElement[] = [];
const listeners = new Set<() => void>();

const resizeObservers = new Map<HTMLInputElement, ResizeObserver>();

let target: HTMLInputElement | null = null;
let listening = false;
let attributeObserver: MutationObserver | null = null;

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function emit() {
  for (const listener of listeners) {
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

function firstFocusable() {
  return fields.filter(isFocusable).sort(byDocumentOrder)[0] ?? null;
}

function isSearchShortcut(event: KeyboardEvent) {
  if (event.repeat || event.altKey || event.shiftKey) {
    return false;
  }
  if (event.key.toLowerCase() !== "f") {
    return false;
  }
  return event.metaKey || event.ctrlKey;
}

function refreshTarget() {
  const next = firstFocusable();
  if (next === target) {
    return;
  }
  target = next;
  emit();
}

function onKeyDown(event: KeyboardEvent) {
  if (!isSearchShortcut(event)) {
    return;
  }
  refreshTarget();
  if (target == null) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  if (document.activeElement === target) {
    target.select();
    return;
  }
  target.focus();
}

// Mount and unmount update the target, but a field can stay mounted while it
// becomes disabled, inert, or hidden. Watch those changes so the hint follows
// the field that can still be focused.
function syncObservers() {
  for (const element of fields) {
    if (resizeObservers.has(element) || typeof ResizeObserver === "undefined") {
      continue;
    }
    const observer = new ResizeObserver(() => {
      refreshTarget();
    });
    observer.observe(element);
    resizeObservers.set(element, observer);
  }
  for (const [element, observer] of resizeObservers) {
    if (fields.includes(element)) {
      continue;
    }
    observer.disconnect();
    resizeObservers.delete(element);
  }

  if (typeof document === "undefined" || typeof MutationObserver === "undefined") {
    return;
  }
  if (fields.length > 0 && attributeObserver == null) {
    attributeObserver = new MutationObserver(() => {
      refreshTarget();
    });
    attributeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["disabled", "hidden", "inert"],
      subtree: true,
    });
    return;
  }
  if (fields.length === 0 && attributeObserver != null) {
    attributeObserver.disconnect();
    attributeObserver = null;
  }
}

function syncListener() {
  if (typeof window === "undefined") {
    return;
  }
  if (fields.length > 0 && !listening) {
    // Capture runs before a menu can stop the key on the way up.
    window.addEventListener("keydown", onKeyDown, true);
    listening = true;
    return;
  }
  if (fields.length === 0 && listening) {
    window.removeEventListener("keydown", onKeyDown, true);
    listening = false;
  }
}

function publish() {
  syncObservers();
  refreshTarget();
  syncListener();
  emit();
}

export function useSearchFieldShortcut(element: HTMLInputElement | null) {
  const isTarget = useSyncExternalStore(
    subscribe,
    () => element != null && target === element,
    () => false,
  );

  useEffect(() => {
    if (element == null) {
      return;
    }
    fields.push(element);
    publish();
    return () => {
      const index = fields.indexOf(element);
      if (index >= 0) {
        fields.splice(index, 1);
      }
      publish();
    };
  }, [element]);

  return isTarget;
}
