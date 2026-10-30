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
import { type MouseEvent, useRef, useState } from "react";

import { Popover, type PopoverProps } from "../Popover/Popover";
import { PopoverPopup } from "../Popover/PopoverPopup";
import { PopoverTrigger } from "../Popover/PopoverTrigger";
import { Text } from "../typography/Text";

import { formatDateRange, orderedDateRange } from "./dateModel";
import { DateMonthView } from "./DateMonthView";
import { isNestedSelectDismiss } from "./datePopover";
import { dateField } from "./variants";

export interface DateRange {
  // Inclusive ISO `YYYY-MM-DD`, or `""` when open.
  from: string;
  to: string;
}

export interface DateRangeFieldProps {
  "className"?: string;
  "id"?: string;
  "name"?: string;
  "value": DateRange;
  "locale"?: string;
  "placeholder"?: string;
  // Shown under the calendar. Explains the two-click selection.
  "hint"?: string;
  // Accessible name for the clear control. Callers pass a translated string.
  "clearLabel"?: string;
  "size"?: 1 | 2;
  "variant"?: "classic" | "surface";
  "disabled"?: boolean;
  "nullable"?: boolean;
  "required"?: boolean;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
  "aria-required"?: boolean;
  "aria-label"?: string;
  "onValueChange"?: (value: DateRange) => void;
}

// One control for a start and end date. The first click sets the start, the
// second sets the end (and swaps them when the second day is earlier). Closing
// the calendar with only a start keeps that open-ended span.
export function DateRangeField(props: DateRangeFieldProps) {
  const {
    className,
    id,
    name,
    value,
    locale,
    placeholder,
    hint,
    clearLabel = "Clear dates",
    size,
    variant,
    disabled = false,
    nullable = false,
    required = false,
    "aria-invalid": ariaInvalid,
    "aria-describedby": ariaDescribedBy,
    "aria-required": ariaRequired,
    "aria-label": ariaLabel,
    onValueChange,
  } = props;

  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<DateRange | null>(null);
  const [preview, setPreview] = useState("");
  const skipCloseCommit = useRef(false);
  const committed = orderedDateRange(value.from, value.to);
  const current = draft ?? committed;
  const label = formatDateRange(current.from, current.to, locale);
  const filled = label !== "";
  const {
    root, surface, glyph, rangeButton, caption, clear, popup, nativeValue,
  } = dateField({ size, variant, filled });

  function publish(next: DateRange) {
    const ordered = orderedDateRange(next.from, next.to);
    if (ordered.from === value.from && ordered.to === value.to) {
      return;
    }
    onValueChange?.(ordered);
  }

  function handleOpenChange(
    nextOpen: boolean,
    eventDetails: Parameters<NonNullable<PopoverProps["onOpenChange"]>>[1],
  ) {
    if (!nextOpen && isNestedSelectDismiss(eventDetails)) {
      return;
    }
    if (nextOpen) {
      skipCloseCommit.current = false;
    }
    if (!nextOpen) {
      if (skipCloseCommit.current) {
        skipCloseCommit.current = false;
      } else if (draft != null) {
        publish(draft);
      }
      setDraft(null);
      setPreview("");
    }
    setOpen(nextOpen);
  }

  function selectDay(iso: string) {
    if (current.from === "" || current.to !== "") {
      setDraft({ from: iso, to: "" });
      setPreview("");
      return;
    }
    const next = iso < current.from
      ? { from: iso, to: current.from }
      : { from: current.from, to: iso };
    skipCloseCommit.current = true;
    setDraft(null);
    setPreview("");
    publish(next);
    setOpen(false);
  }

  function hoverDay(iso: string) {
    if (current.to !== "" || current.from === "") {
      if (preview !== "") {
        setPreview("");
      }
      return;
    }
    setPreview(iso);
  }

  function clearRange(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    event.stopPropagation();
    skipCloseCommit.current = true;
    setDraft(null);
    setPreview("");
    publish({ from: "", to: "" });
  }

  const hasValue = value.from !== "" || value.to !== "" || draft != null;

  return (
    <div className={root({ className })}>
      {(name != null || required) && (
        <>
          <input
            type="text"
            className={nativeValue()}
            tabIndex={-1}
            aria-hidden
            name={name == null ? undefined : `${name}From`}
            value={value.from}
            required={required}
            disabled={disabled}
            onChange={() => undefined}
          />
          <input
            type="text"
            className={nativeValue()}
            tabIndex={-1}
            aria-hidden
            name={name == null ? undefined : `${name}To`}
            value={value.to}
            required={required}
            disabled={disabled}
            onChange={() => undefined}
          />
        </>
      )}
      <div className={surface()}>
        <Popover open={open} onOpenChange={handleOpenChange}>
          <PopoverTrigger
            id={id}
            disabled={disabled}
            className={rangeButton()}
            aria-invalid={ariaInvalid}
            aria-describedby={ariaDescribedBy}
            aria-required={ariaRequired || required || undefined}
            aria-label={ariaLabel}
          >
            <span className={glyph()} aria-hidden>
              <CalendarBlankIcon />
            </span>
            <span className={caption()}>{filled ? label : placeholder}</span>
          </PopoverTrigger>
          <PopoverPopup>
            <div className={popup()}>
              <DateMonthView
                size={size}
                disabled={disabled}
                value={current.from !== "" ? current.from : current.to}
                locale={locale}
                resetKey={open}
                range={{
                  start: current.from,
                  end: current.to,
                  preview: current.to === "" ? preview : "",
                }}
                onDayHover={hoverDay}
                onDaySelect={selectDay}
              />
              {hint != null && hint !== "" && (
                <Text size={1} color="faint">{hint}</Text>
              )}
            </div>
          </PopoverPopup>
        </Popover>
        {nullable && hasValue && !disabled && (
          <button
            type="button"
            className={clear()}
            aria-label={clearLabel}
            onMouseDown={(event) => {
              event.preventDefault();
              clearRange(event);
            }}
            onClick={clearRange}
          >
            <XIcon />
          </button>
        )}
      </div>
    </div>
  );
}
