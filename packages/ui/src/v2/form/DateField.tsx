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

import { CalendarBlankIcon, XIcon } from "@phosphor-icons/react";
import { useState } from "react";

import { Popover, type PopoverProps } from "../Popover/Popover";
import { PopoverPopup } from "../Popover/PopoverPopup";
import { PopoverTrigger } from "../Popover/PopoverTrigger";

import { DateMonthView } from "./DateMonthView";
import { useDateInput } from "./useDateInput";
import { dateField } from "./variants";

export interface DateFieldProps {
  // Applied to the trigger row (the top-level element).
  "className"?: string;
  // Forwards onto the typed input so Field can associate its label.
  "id"?: string;
  // Native form field name; written to a hidden input.
  "name"?: string;
  // ISO `YYYY-MM-DD`, or `""` when empty.
  "value": string;
  // `Intl` locale for typed mask, month names, and weekdays.
  "locale"?: string;
  // Shown on the input when empty. Defaults to the locale mask (`mm/dd/yyyy`).
  "placeholder"?: string;
  "size"?: 1 | 2;
  // Surface treatment (defaults to "surface").
  "variant"?: "classic" | "surface";
  "disabled"?: boolean;
  // Inclusive ISO bound; days before this are not selectable.
  "min"?: string;
  // Inclusive ISO bound; days after this are not selectable.
  "max"?: string;
  // When true, a clear control appears while a date is selected.
  "nullable"?: boolean;
  "required"?: boolean;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
  "aria-required"?: boolean;
  "onValueChange"?: (value: string) => void;
}

// TextField-like date control: type digits and separators are inserted, or
// pick a day from the Popover calendar. Value is always an ISO date string.
export function DateField(props: DateFieldProps) {
  const {
    className,
    id,
    name,
    value,
    locale,
    placeholder,
    size,
    variant,
    disabled = false,
    min,
    max,
    nullable = false,
    required = false,
    "aria-invalid": ariaInvalid,
    "aria-describedby": ariaDescribedBy,
    "aria-required": ariaRequired,
    onValueChange,
  } = props;

  const [open, setOpen] = useState(false);
  const { root, surface, iconTrigger, input, clear } = dateField({ size, variant });
  const dateInput = useDateInput({
    value,
    locale,
    min,
    max,
    nullable,
    onCommit: iso => onValueChange?.(iso),
  });

  function handleOpenChange(
    nextOpen: boolean,
    eventDetails: Parameters<NonNullable<PopoverProps["onOpenChange"]>>[1],
  ) {
    if (!nextOpen && isNestedSelectDismiss(eventDetails)) {
      return;
    }
    setOpen(nextOpen);
  }

  function selectDay(iso: string) {
    onValueChange?.(iso);
    setOpen(false);
  }

  return (
    <div className={root({ className })}>
      {name != null && <input type="hidden" name={name} value={value} disabled={disabled} />}
      <div className={surface()}>
        <Popover open={open} onOpenChange={handleOpenChange}>
          <PopoverTrigger
            disabled={disabled}
            className={iconTrigger()}
            aria-label="Choose date"
          >
            <CalendarBlankIcon />
          </PopoverTrigger>
          <PopoverPopup>
            <DateMonthView
              size={size}
              disabled={disabled}
              value={value}
              locale={locale}
              min={min}
              max={max}
              resetKey={open}
              onDaySelect={selectDay}
            />
          </PopoverPopup>
        </Popover>
        <input
          id={id}
          type="text"
          inputMode="numeric"
          autoComplete="off"
          spellCheck={false}
          disabled={disabled}
          required={required}
          maxLength={10}
          placeholder={placeholder ?? dateInput.placeholder}
          value={dateInput.draft}
          aria-invalid={ariaInvalid}
          aria-describedby={ariaDescribedBy}
          aria-required={ariaRequired}
          className={input()}
          onChange={dateInput.onChange}
          onBlur={dateInput.onBlur}
          onKeyDown={dateInput.onKeyDown}
        />
        {nullable && value !== "" && !disabled && (
          <button
            type="button"
            className={clear()}
            aria-label="Clear date"
            onMouseDown={event => event.preventDefault()}
            onClick={() => onValueChange?.("")}
          >
            <XIcon />
          </button>
        )}
      </div>
    </div>
  );
}

function isNestedSelectDismiss(
  eventDetails: Parameters<NonNullable<PopoverProps["onOpenChange"]>>[1],
): boolean {
  if (eventDetails.reason === "focus-out") {
    const relatedTarget = eventDetails.event instanceof FocusEvent
      ? eventDetails.event.relatedTarget
      : null;
    return relatedTarget instanceof Element
      && relatedTarget.closest("[role='listbox'], [role='option']") != null;
  }
  if (eventDetails.reason !== "outside-press") {
    return false;
  }
  const target = eventDetails.event.target;
  return target instanceof Element
    && target.closest("[role='listbox'], [role='option']") != null;
}
