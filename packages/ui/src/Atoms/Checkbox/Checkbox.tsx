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
