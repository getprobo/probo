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

import { DateField } from "./DateField";
import { DateFieldSkeleton } from "./DateFieldSkeleton";
import { Field } from "./Field";

export default {
  title: "v2/form/DateField",
  component: DateField,
};

export function Default() {
  return (
    <div className="w-60">
      <DateField value="" />
    </div>
  );
}

export function Controlled() {
  const [value, setValue] = useState("2026-09-24");
  return (
    <div className="flex flex-col gap-3">
      <div className="w-60">
        <DateField
          value={value}
          nullable
          onValueChange={setValue}
        />
      </div>
      <span className="text-2 text-sand-11">
        Value:
        {value || "(empty)"}
      </span>
    </div>
  );
}

export function Sizes() {
  return (
    <div className="flex w-60 flex-col gap-3">
      <DateField size={1} value="2026-03-12" />
      <DateField size={2} value="2026-03-12" />
    </div>
  );
}

export function MinMax() {
  const [value, setValue] = useState("2026-09-15");
  return (
    <div className="w-60">
      <Field label="Between the 10th and the 20th">
        <DateField
          value={value}
          min="2026-09-10"
          max="2026-09-20"
          onValueChange={setValue}
        />
      </Field>
    </div>
  );
}

export function Typed() {
  const [value, setValue] = useState("");
  return (
    <div className="flex flex-col gap-3">
      <div className="w-60">
        <Field label="Type 09242026">
          <DateField
            value={value}
            locale="en-US"
            nullable
            onValueChange={setValue}
          />
        </Field>
      </div>
      <span className="text-2 text-sand-11">
        Value:
        {value || "(empty)"}
      </span>
    </div>
  );
}

export function Skeleton() {
  return <DateFieldSkeleton />;
}
