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

import { XIcon } from "@phosphor-icons/react";

import { DateMonthView } from "./DateMonthView";
import { useDateInput } from "./useDateInput";
import { dateCalendar } from "./variants";

export interface DateCalendarProps {
  // Applied to the panel (the top-level element).
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
  // Surface treatment of the typed input (defaults to "surface").
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

// Calendar panel with a typed input above the month grid. Value is always an
// ISO date string. No popover — the calendar is always visible.
export function DateCalendar(props: DateCalendarProps) {
  const {
    className,
    id,
    name,
    value,
    locale,
    placeholder,
    size = 2,
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

  const { root, surface, input, clear } = dateCalendar({ size, variant });
  const dateInput = useDateInput({
    value,
    locale,
    min,
    max,
    nullable,
    onCommit: iso => onValueChange?.(iso),
  });

  return (
    <div className={root({ className })}>
      {name != null && <input type="hidden" name={name} value={value} disabled={disabled} />}
      <div className={surface()}>
        <input
          id={id}
          type="text"
          inputMode="numeric"
          autoComplete="off"
          spellCheck={false}
          disabled={disabled}
          required={required}
          pattern={dateInput.matchesValue ? undefined : "a^"}
          maxLength={dateInput.placeholder.length}
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
      <DateMonthView
        value={value}
        locale={locale}
        min={min}
        max={max}
        size={size}
        disabled={disabled}
        onDaySelect={iso => onValueChange?.(iso)}
      />
    </div>
  );
}
