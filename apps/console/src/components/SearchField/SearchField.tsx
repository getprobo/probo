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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { TextField, type TextFieldProps } from "@probo/ui/src/v2/form/TextField";
import { shortcutAria } from "@probo/ui/src/v2/shortcutAria";
import { useCallback, useState } from "react";

import { SearchShortcutHint } from "#/components/SearchShortcutHint/SearchShortcutHint";

import { useFocusTarget } from "./useFocusTarget";

export type SearchFieldProps = Omit<TextFieldProps, "icon" | "ref" | "suffix">;

function fieldText(value: TextFieldProps["value"]) {
  if (value == null) {
    return "";
  }
  return String(value);
}

export function SearchField(props: SearchFieldProps) {
  const { defaultValue, onValueChange, value, ...rest } = props;
  const [input, setInput] = useState<HTMLInputElement | null>(null);
  const isTarget = useFocusTarget(input, "f");
  const [uncontrolledValue, setUncontrolledValue] = useState(() => fieldText(defaultValue));
  // `value === undefined` matches Base UI: null is still a controlled value.
  const text = value === undefined ? uncontrolledValue : fieldText(value);
  const showHint = isTarget && text.length === 0;

  const setNode = useCallback((node: HTMLElement | null) => {
    setInput(node instanceof HTMLInputElement ? node : null);
  }, []);

  const handleValueChange: NonNullable<TextFieldProps["onValueChange"]> = (next, eventDetails) => {
    if (value === undefined) {
      setUncontrolledValue(next);
    }
    onValueChange?.(next, eventDetails);
  };

  return (
    <TextField
      {...rest}
      ref={setNode}
      icon={<MagnifyingGlassIcon />}
      defaultValue={defaultValue}
      value={value}
      onValueChange={handleValueChange}
      suffix={showHint ? <SearchShortcutHint shortcut="F" /> : undefined}
      aria-keyshortcuts={isTarget ? shortcutAria("F") : undefined}
    />
  );
}
