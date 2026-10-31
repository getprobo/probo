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

import { useState } from "react";

import { type DateRange, DateRangeField } from "./DateRangeField";
import { DateRangeFieldSkeleton } from "./DateRangeFieldSkeleton";

export default {
  title: "v2/form/DateRangeField",
  component: DateRangeField,
};

const EMPTY: DateRange = { from: "", to: "" };

export function Default() {
  return (
    <div className="w-64">
      <DateRangeField
        value={EMPTY}
        nullable
        placeholder="Date range"
        hint="Select a start date, then an end date."
      />
    </div>
  );
}

export function Controlled() {
  const [value, setValue] = useState<DateRange>({ from: "2026-09-10", to: "2026-09-24" });
  return (
    <div className="flex w-64 flex-col gap-3">
      <DateRangeField
        value={value}
        nullable
        placeholder="Date range"
        hint="Select a start date, then an end date."
        onValueChange={setValue}
      />
      <span className="text-2 text-sand-11">
        {value.from || "(empty)"}
        {" – "}
        {value.to || "(empty)"}
      </span>
    </div>
  );
}

export function Skeleton() {
  return (
    <div className="w-64">
      <DateRangeFieldSkeleton />
    </div>
  );
}
