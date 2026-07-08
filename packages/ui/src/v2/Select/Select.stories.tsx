// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
        {value ?? "none"}
      </span>
    </div>
  );
}

export function Skeleton() {
  return <SelectSkeleton />;
}
