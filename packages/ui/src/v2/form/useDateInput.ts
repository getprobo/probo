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

import { type ChangeEvent, type KeyboardEvent, useState } from "react";

import {
  datePattern,
  digitsToIso,
  isoToMasked,
  isOutOfRange,
  maskDigits,
  maskPlaceholder,
  parseIsoDate,
} from "./dateModel";

export function useDateInput({
  value,
  locale,
  min,
  max,
  nullable = false,
  onCommit,
}: {
  value: string;
  locale?: string;
  min?: string;
  max?: string;
  nullable?: boolean;
  onCommit: (iso: string) => void;
}) {
  const pattern = datePattern(locale);
  const [draft, setDraft] = useState(() => isoToMasked(value, pattern));
  const [source, setSource] = useState({ value, locale });

  if (value !== source.value || locale !== source.locale) {
    setSource({ value, locale });
    setDraft(isoToMasked(value, pattern));
  }

  const minDate = parseIsoDate(min ?? "");
  const maxDate = parseIsoDate(max ?? "");

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const raw = event.currentTarget.value;
    const trimmed = raw.trim();
    if (parseIsoDate(trimmed) != null && !isOutOfRange(trimmed, minDate, maxDate)) {
      onCommit(trimmed);
      return;
    }

    const digits = raw.replace(/\D/g, "").slice(0, 8);
    setDraft(maskDigits(digits, pattern));

    if (digits.length === 0) {
      if (nullable) {
        onCommit("");
      }
      return;
    }

    if (digits.length !== 8) {
      return;
    }

    const iso = digitsToIso(digits, pattern);
    if (iso != null && !isOutOfRange(iso, minDate, maxDate)) {
      onCommit(iso);
    }
  }

  function handleBlur() {
    setDraft(isoToMasked(value, pattern));
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Backspace") {
      return;
    }
    const field = event.currentTarget;
    const start = field.selectionStart ?? 0;
    const end = field.selectionEnd ?? 0;
    if (start !== end || start === 0) {
      return;
    }
    const charBefore = draft[start - 1];
    if (charBefore == null || /\d/.test(charBefore)) {
      return;
    }
    event.preventDefault();
    const prefixDigits = draft.slice(0, start).replace(/\D/g, "").slice(0, -1);
    const suffixDigits = draft.slice(start).replace(/\D/g, "");
    const digits = `${prefixDigits}${suffixDigits}`.slice(0, 8);
    setDraft(maskDigits(digits, pattern));
    if (digits.length === 0 && nullable) {
      onCommit("");
    }
  }

  return {
    draft,
    placeholder: maskPlaceholder(pattern),
    onChange: handleChange,
    onBlur: handleBlur,
    onKeyDown: handleKeyDown,
  };
}
