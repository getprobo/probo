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

import type { PopoverProps } from "../Popover/Popover";

// Month and year selects inside the calendar portal their listbox to the body.
// That press is outside the date popover, but it belongs to the same control.
export function isNestedSelectDismiss(
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
