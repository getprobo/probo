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

import { ToggleGroup as BaseToggleGroup } from "@base-ui/react/toggle-group";
import type { ReactNode } from "react";

import { segmentedControl } from "./variants";

export type SegmentedControlProps = {
  // Single selected value (controlled).
  value?: string;
  // Single selected value (uncontrolled).
  defaultValue?: string;
  // Fired with the newly selected value. Never fired with an empty selection,
  // so a value always stays selected (clicking the active item is a no-op).
  onValueChange?: (value: string) => void;
  disabled?: boolean;
  className?: string;
  children?: ReactNode;
};

// Single-select pill group. Wraps Base UI's array-based ToggleGroup with a
// friendlier single-value API.
export function SegmentedControl(props: SegmentedControlProps) {
  const { value, defaultValue, onValueChange, disabled, className, children } = props;
  const { root } = segmentedControl();

  return (
    <BaseToggleGroup
      className={root({ className })}
      disabled={disabled}
      value={value != null ? [value] : undefined}
      defaultValue={defaultValue != null ? [defaultValue] : undefined}
      onValueChange={(groupValue) => {
        const next = groupValue[0];
        if (next != null) {
          onValueChange?.(next);
        }
      }}
    >
      {children}
    </BaseToggleGroup>
  );
}
