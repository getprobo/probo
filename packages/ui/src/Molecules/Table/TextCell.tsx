// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
