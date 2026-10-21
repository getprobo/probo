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

import { useEffect, useRef } from "react";

export interface UseHotkeysOptions {
  preventDefault?: boolean;
  enabled?: boolean;
  capture?: boolean;
}

interface HotkeySlot {
  key: string;
  capture: boolean;
  preventDefault: boolean;
  handler: (event: KeyboardEvent) => void;
  handlerRef: { current: (event: KeyboardEvent) => void };
  count: number;
}

const slots: HotkeySlot[] = [];

let captureListening = false;
let bubbleListening = false;

function usesMetaShortcut() {
  if (typeof navigator === "undefined") {
    return false;
  }
  return /Mac|iPhone|iPad|iPod/.test(navigator.userAgent);
}

function parseCombo(combo: string) {
  const match = /^mod\+([a-z])$/i.exec(combo);
  if (match == null) {
    throw new Error(`unsupported hotkey: ${combo}`);
  }
  return match[1].toLowerCase();
}

function matches(event: KeyboardEvent, key: string) {
  if (event.repeat || event.altKey || event.shiftKey) {
    return false;
  }
  if (event.key.toLowerCase() !== key) {
    return false;
  }
  return usesMetaShortcut() ? event.metaKey : event.ctrlKey;
}

function onKeyDown(event: KeyboardEvent) {
  const capture = event.eventPhase === Event.CAPTURING_PHASE;
  for (const slot of slots) {
    if (slot.capture !== capture || !matches(event, slot.key)) {
      continue;
    }
    if (slot.preventDefault) {
      event.preventDefault();
    }
    slot.handlerRef.current(event);
  }
}

function syncListeners() {
  if (typeof window === "undefined") {
    return;
  }
  const needsCapture = slots.some(slot => slot.capture);
  const needsBubble = slots.some(slot => !slot.capture);
  if (needsCapture && !captureListening) {
    window.addEventListener("keydown", onKeyDown, true);
    captureListening = true;
  }
  if (!needsCapture && captureListening) {
    window.removeEventListener("keydown", onKeyDown, true);
    captureListening = false;
  }
  if (needsBubble && !bubbleListening) {
    window.addEventListener("keydown", onKeyDown);
    bubbleListening = true;
  }
  if (!needsBubble && bubbleListening) {
    window.removeEventListener("keydown", onKeyDown);
    bubbleListening = false;
  }
}

function bindHotkey(
  combo: string,
  handler: (event: KeyboardEvent) => void,
  handlerRef: { current: (event: KeyboardEvent) => void },
  options?: UseHotkeysOptions,
) {
  const key = parseCombo(combo);
  const capture = options?.capture ?? false;
  const preventDefault = options?.preventDefault ?? false;
  const existing = slots.find(slot =>
    slot.handler === handler
    && slot.key === key
    && slot.capture === capture
    && slot.preventDefault === preventDefault,
  );
  if (existing != null) {
    existing.count += 1;
    existing.handlerRef = handlerRef;
    return () => {
      release(existing);
    };
  }

  const slot: HotkeySlot = {
    key,
    capture,
    preventDefault,
    handler,
    handlerRef,
    count: 1,
  };
  slots.push(slot);
  syncListeners();
  return () => {
    release(slot);
  };
}

function release(slot: HotkeySlot) {
  slot.count -= 1;
  if (slot.count > 0) {
    return;
  }
  const index = slots.indexOf(slot);
  if (index >= 0) {
    slots.splice(index, 1);
  }
  syncListeners();
}

export function useHotkeys(
  combo: string,
  handler: (event: KeyboardEvent) => void,
  options?: UseHotkeysOptions,
) {
  const handlerRef = useRef(handler);
  const preventDefault = options?.preventDefault ?? false;
  const enabled = options?.enabled ?? true;
  const capture = options?.capture ?? false;

  useEffect(() => {
    handlerRef.current = handler;
  }, [handler]);

  useEffect(() => {
    if (!enabled) {
      return;
    }
    // handlerRef.current is the dedupe identity. A fresh callback keeps this
    // binding and is read from the ref; a stable shared handler refcounts.
    return bindHotkey(combo, handlerRef.current, handlerRef, {
      preventDefault,
      capture,
    });
  }, [capture, combo, enabled, preventDefault]);
}
