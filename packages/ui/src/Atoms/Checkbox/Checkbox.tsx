import { useState } from "react";
import { tv } from "tailwind-variants";

import { IconCheckmark1 } from "../Icons";

type Props = {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
};

const checkbox = tv({
  base: "size-4 border border-border-mid relative rounded-sm flex items-center justify-center",
  variants: {
    isFocused: {
      true: "shadow shadow-focus",
    },
    checked: {
      true: "bg-accent text-invert",
    },
    disabled: {
      true: "opacity-50 cursor-not-allowed",
    },
  },
});

export function Checkbox({ checked, onChange, disabled = false }: Props) {
  const [isFocused, setFocus] = useState(false);
  return (
    <div className={checkbox({ isFocused, checked, disabled })}>
      <input
        className="absolute inset-0 opacity-0"
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={e => !disabled && onChange(e.target.checked)}
        onFocus={() => !disabled && setFocus(true)}
        onBlur={() => setFocus(false)}
      />
      {checked && <IconCheckmark1 size={12} />}
    </div>
  );
}
