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

type HotkeyHandler = (event: KeyboardEvent) => void;

interface HotkeySlot {
  key: string;
  capture: boolean;
  preventDefault: boolean;
  handlerRef: { current: HotkeyHandler };
}

const slots: HotkeySlot[] = [];

let captureListening = false;
let bubbleListening = false;

export function usesMetaShortcut() {
  if (typeof navigator === "undefined") {
    return false;
  }
  return /Mac|iPhone|iPad|iPod/.test(navigator.userAgent);
}

function parseCombo(combo: string) {
  const match = /^mod\+(enter|[a-z.])$/i.exec(combo);
  if (match == null) {
    throw new Error(`unsupported hotkey: ${combo}`);
  }
  return match[1].toLowerCase();
}

function matches(event: KeyboardEvent, key: string) {
  if (event.repeat || event.altKey) {
    return false;
  }
  // Period is Shift+; on AZERTY, so Command-Period arrives with shiftKey set.
  if (event.shiftKey && key !== ".") {
    return false;
  }
  if (event.key.toLowerCase() !== key) {
    return false;
  }
  if (usesMetaShortcut()) {
    return event.metaKey && !event.ctrlKey;
  }
  return event.ctrlKey && !event.metaKey;
}

function onKeyDown(event: KeyboardEvent) {
  const capture = event.eventPhase === Event.CAPTURING_PHASE;
  // Several mounts can share one stable callback. Run that function once.
  const seen: { handler: HotkeyHandler; preventDefault: boolean }[] = [];
  for (const slot of slots) {
    if (slot.capture !== capture || !matches(event, slot.key)) {
      continue;
    }
    const handler = slot.handlerRef.current;
    const duplicate = seen.some(item =>
      item.handler === handler && item.preventDefault === slot.preventDefault,
    );
    if (duplicate) {
      continue;
    }
    seen.push({ handler, preventDefault: slot.preventDefault });
    if (slot.preventDefault) {
      event.preventDefault();
    }
    handler(event);
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
  handlerRef: { current: HotkeyHandler },
  options?: UseHotkeysOptions,
) {
  const key = parseCombo(combo);
  const capture = options?.capture ?? false;
  const preventDefault = options?.preventDefault ?? false;
  const slot: HotkeySlot = {
    key,
    capture,
    preventDefault,
    handlerRef,
  };
  slots.push(slot);
  syncListeners();
  return () => {
    const index = slots.indexOf(slot);
    if (index >= 0) {
      slots.splice(index, 1);
    }
    syncListeners();
  };
}

export function useHotkeys(
  combo: string,
  handler: HotkeyHandler,
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
    return bindHotkey(combo, handlerRef, {
      preventDefault,
      capture,
    });
  }, [capture, combo, enabled, preventDefault]);
}
