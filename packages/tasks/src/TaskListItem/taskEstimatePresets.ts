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

import { stringifyTaskDuration, taskDurationsEqual } from "../taskDuration";

export const noneEstimateValue = "__NONE__";

export const estimatePresets = [
  stringifyTaskDuration(15, "M"),
  stringifyTaskDuration(30, "M"),
  stringifyTaskDuration(1, "H"),
  stringifyTaskDuration(2, "H"),
  stringifyTaskDuration(4, "H"),
  stringifyTaskDuration(1, "D"),
  stringifyTaskDuration(1, "W"),
].filter((value): value is string => value != null);

export function matchingEstimatePreset(value?: string | null): string | null {
  if (value == null) {
    return null;
  }
  return estimatePresets.find(preset => taskDurationsEqual(preset, value)) ?? value;
}
