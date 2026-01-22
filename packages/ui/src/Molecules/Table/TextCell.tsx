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
