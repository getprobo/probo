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

import { ArrowElbowDownLeftIcon, CommandIcon, ControlIcon } from "@phosphor-icons/react";

import type { ButtonProps } from "../Button/Button";
import { usesMetaShortcut } from "../shortcutAria";

import { shortcutHint } from "./variants";

export interface ShortcutHintProps {
  shortcut: string;
  chip?: boolean;
  variant: "soft" | "solid";
  color: NonNullable<ButtonProps["color"]>;
}

export function ShortcutHint({ shortcut, chip = false, variant, color }: ShortcutHintProps) {
  const { hint, modifier } = shortcutHint({ chip, variant, color });
  const ModifierIcon = usesMetaShortcut() ? CommandIcon : ControlIcon;
  const key = shortcut.toLowerCase();

  return (
    <span data-shortcut-hint="" aria-hidden="true" className={hint()}>
      <ModifierIcon className={modifier()} />
      {key === "enter"
        ? <ArrowElbowDownLeftIcon className={modifier()} />
        : shortcut.toUpperCase()}
    </span>
  );
}
