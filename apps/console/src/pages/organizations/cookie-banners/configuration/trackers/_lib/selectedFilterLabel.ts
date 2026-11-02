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

import {
  type DetectedTrackerObserver,
  isDetectedTrackerObserver,
} from "./useTrackersListFilters";

export const allObserversValue = "ALL";

// Clicking All clears a specific selection. Clicking an observer while All
// is checked keeps that observer and drops All.
export function observedByFromMenu(
  previous: readonly DetectedTrackerObserver[],
  next: readonly string[],
): DetectedTrackerObserver[] {
  const observers = next.filter(isDetectedTrackerObserver);
  if (next.includes(allObserversValue) && previous.length > 0) {
    return [];
  }
  return observers;
}

export function selectedFilterLabel(
  values: readonly string[],
  emptyLabel: string,
  label: (value: DetectedTrackerObserver) => string,
  countLabel: string,
): string {
  if (values.length === 0) {
    return emptyLabel;
  }
  const first = values[0];
  if (values.length === 1 && first != null && isDetectedTrackerObserver(first)) {
    return label(first);
  }
  return countLabel;
}
