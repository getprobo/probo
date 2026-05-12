// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { type KeyboardEventHandler, useRef, useState } from "react";

import { EditableCell, useEditableCellRef } from "./EditableCell";
import { useEditableRowContext } from "./EditableRow";

type Props = {
  name: string;
  defaultValue: string;
  required?: boolean;
};

export function TextCell(props: Props) {
  const [value, setValue] = useState(props.defaultValue);
  const inputRef = useRef<HTMLInputElement>(null);
  const cellRef = useEditableCellRef();
  const blurOnTab: KeyboardEventHandler<HTMLInputElement> = (e) => {
    if (e.key === "Tab") {
      cellRef.current?.close();
    }
  };
  const { onUpdate } = useEditableRowContext();
  const onClose = () => {
    const inputValue = (inputRef.current?.value ?? "").trim();
    // Do not propagate empty value for required fields
    if (props.required && inputValue === "") {
      return;
    }
    if (inputValue !== value) {
      setValue(inputValue);
      onUpdate(props.name, inputValue);
    }
  };

  return (
    <EditableCell
      name={props.name}
      label={value}
      ref={cellRef}
      onClose={onClose}
    >
      <input
        type="text"
        ref={inputRef}
        defaultValue={props.defaultValue}
        onKeyDown={blurOnTab}
        className="text-sm text-txt-primary outline-none"
        style={{ paddingLeft: "var(--padding)" }}
      />
    </EditableCell>
  );
}
