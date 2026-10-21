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

import { usesMetaShortcut } from "@probo/hotkeys";
import { useEffect, useState } from "react";

function shortcutModifierDown(event: KeyboardEvent) {
  return usesMetaShortcut() ? event.metaKey : event.ctrlKey;
}

function isShortcutModifier(event: KeyboardEvent) {
  return usesMetaShortcut() ? event.key === "Meta" : event.key === "Control";
}

export function useModifierPressed() {
  const [pressed, setPressed] = useState(false);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.repeat) {
        return;
      }
      if (isShortcutModifier(event) || shortcutModifierDown(event)) {
        setPressed(true);
      }
    }

    function onKeyUp(event: KeyboardEvent) {
      if (!isShortcutModifier(event)) {
        return;
      }
      setPressed(shortcutModifierDown(event));
    }

    function clear() {
      setPressed(false);
    }

    function onVisibilityChange() {
      if (document.visibilityState === "hidden") {
        clear();
      }
    }

    window.addEventListener("keydown", onKeyDown, true);
    window.addEventListener("keyup", onKeyUp, true);
    window.addEventListener("blur", clear);
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () => {
      window.removeEventListener("keydown", onKeyDown, true);
      window.removeEventListener("keyup", onKeyUp, true);
      window.removeEventListener("blur", clear);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, []);

  return pressed;
}
