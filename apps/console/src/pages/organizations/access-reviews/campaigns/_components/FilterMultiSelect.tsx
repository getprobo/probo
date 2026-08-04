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

import { Checkbox, IconChevronDown } from "@probo/ui";
import * as Popover from "@radix-ui/react-popover";
import { useState } from "react";

export type FilterMultiSelectOption = {
  value: string;
  label: string;
};

type Props = {
  placeholder: string;
  options: ReadonlyArray<FilterMultiSelectOption>;
  value: ReadonlyArray<string>;
  onChange: (value: string[]) => void;
  onOpenChange?: (open: boolean) => void;
  side?: "top" | "bottom";
  className?: string;
  disabled?: boolean;
};

// Multi-select control matching the subprocessors toolbar trigger size.
// Empty selection shows the placeholder — there is no synthetic "All" option.
export function FilterMultiSelect({
  placeholder,
  options,
  value,
  onChange,
  onOpenChange,
  side = "bottom",
  className,
  disabled = false,
}: Props) {
  const [open, setOpen] = useState(false);

  const selectedLabels = options
    .filter(option => value.includes(option.value))
    .map(option => option.label);

  const triggerLabel = value.length === 0
    ? placeholder
    : value.length === 1
      ? (selectedLabels[0] ?? placeholder)
      : `${value.length} selected`;

  const toggle = (optionValue: string) => {
    if (value.includes(optionValue)) {
      onChange(value.filter(v => v !== optionValue));
    } else {
      onChange([...value, optionValue]);
    }
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (disabled) {
      return;
    }
    setOpen(nextOpen);
    onOpenChange?.(nextOpen);
  };

  return (
    <Popover.Root open={open} onOpenChange={handleOpenChange}>
      <Popover.Trigger asChild>
        <button
          type="button"
          disabled={disabled}
          className={[
            "flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-border-low bg-level-1 px-3 text-sm text-txt-primary hover:bg-tertiary-hover disabled:cursor-not-allowed disabled:opacity-50",
            className,
          ].filter(Boolean).join(" ")}
        >
          <span className={value.length === 0 ? "truncate text-txt-tertiary" : "truncate"}>
            {triggerLabel}
          </span>
          <IconChevronDown size={14} className="shrink-0 text-txt-tertiary" />
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          align="start"
          side={side}
          sideOffset={4}
          className="z-100 max-h-(--radix-popover-content-available-height) w-(--radix-popover-trigger-width) min-w-48 overflow-y-auto rounded-[10px] border border-border-low bg-level-1 p-1 text-txt-primary shadow-mid animate-in fade-in slide-in-from-top-1"
        >
          {options.map(option => (
            <label
              key={option.value}
              className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-txt-primary hover:bg-tertiary-hover"
            >
              <Checkbox
                checked={value.includes(option.value)}
                onChange={() => toggle(option.value)}
              />
              <span className="truncate text-txt-primary">{option.label}</span>
            </label>
          ))}
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
