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

import { Select } from "./Select";
import { SelectItem } from "./SelectItem";
import { SelectPopup } from "./SelectPopup";
import { SelectSkeleton } from "./SelectSkeleton";
import { SelectTrigger } from "./SelectTrigger";

export default {
  title: "v2/Select",
  component: Select,
};

const fruits: Record<string, string> = {
  apple: "Apple",
  banana: "Banana",
  cherry: "Cherry",
};

export function Default() {
  return (
    <div className="w-40">
      <Select>
        <SelectTrigger placeholder="Select a fruit">
          {(value: string | null) => (value ? fruits[value] : null)}
        </SelectTrigger>
        <SelectPopup>
          {Object.entries(fruits).map(([value, label]) => (
            <SelectItem key={value} value={value}>{label}</SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
}

export function Variants() {
  return (
    <div className="flex flex-col gap-3">
      {(["classic", "surface", "soft", "ghost"] as const).map(variant => (
        <div key={variant} className="w-40">
          <Select>
            <SelectTrigger variant={variant} placeholder={variant}>
              {(value: string | null) => (value ? fruits[value] : null)}
            </SelectTrigger>
            <SelectPopup>
              {Object.entries(fruits).map(([value, label]) => (
                <SelectItem key={value} value={value}>{label}</SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      ))}
    </div>
  );
}

export function Controlled() {
  const [value, setValue] = useState<string | null>(null);
  return (
    <div className="flex flex-col gap-3">
      <div className="w-40">
        <Select value={value} onValueChange={setValue}>
          <SelectTrigger placeholder="All categories">
            {(current: string | null) => (current ? fruits[current] : null)}
          </SelectTrigger>
          <SelectPopup>
            {Object.entries(fruits).map(([key, label]) => (
              <SelectItem key={key} value={key}>{label}</SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
      <span className="text-2 text-sand-11">
        Selected:
        {" "}
        {value ?? "none"}
      </span>
    </div>
  );
}

export function Skeleton() {
  return <SelectSkeleton />;
}
