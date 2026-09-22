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

import { Combobox } from "./Combobox";
import { ComboboxChip } from "./ComboboxChip";
import { ComboboxChipRemove } from "./ComboboxChipRemove";
import { ComboboxChips } from "./ComboboxChips";
import { ComboboxEmpty } from "./ComboboxEmpty";
import { ComboboxInput } from "./ComboboxInput";
import { ComboboxInputGroup } from "./ComboboxInputGroup";
import { ComboboxItem } from "./ComboboxItem";
import { ComboboxList } from "./ComboboxList";
import { ComboboxPopup } from "./ComboboxPopup";
import { ComboboxValue } from "./ComboboxValue";

export default {
  title: "v2/Combobox",
  component: Combobox,
};

type Fruit = { value: string; label: string };

const fruits: Fruit[] = [
  { value: "apple", label: "Apple" },
  { value: "banana", label: "Banana" },
  { value: "cherry", label: "Cherry" },
  { value: "date", label: "Date" },
  { value: "fig", label: "Fig" },
];

export function Single() {
  const [value, setValue] = useState<Fruit | null>(null);

  return (
    <div className="w-72">
      <Combobox
        items={fruits}
        value={value}
        onValueChange={setValue}
        itemToStringLabel={item => item.label}
      >
        <ComboboxInputGroup>
          <ComboboxInput placeholder="Search fruit" />
        </ComboboxInputGroup>
        <ComboboxPopup>
          <ComboboxEmpty>No fruit found</ComboboxEmpty>
          <ComboboxList<Fruit>>
            {fruit => (
              <ComboboxItem key={fruit.value} value={fruit}>{fruit.label}</ComboboxItem>
            )}
          </ComboboxList>
        </ComboboxPopup>
      </Combobox>
    </div>
  );
}

export function Multiple() {
  const [value, setValue] = useState<Fruit[]>([]);

  return (
    <div className="w-96">
      <Combobox
        multiple
        items={fruits}
        value={value}
        onValueChange={setValue}
        itemToStringLabel={item => item.label}
      >
        <ComboboxInputGroup>
          <ComboboxChips>
            <ComboboxValue<Fruit[]>>
              {selected => (
                <>
                  {selected.map(fruit => (
                    <ComboboxChip key={fruit.value} aria-label={fruit.label}>
                      {fruit.label}
                      <ComboboxChipRemove aria-label={`Remove ${fruit.label}`} />
                    </ComboboxChip>
                  ))}
                  <ComboboxInput placeholder={selected.length === 0 ? "Search fruit" : undefined} />
                </>
              )}
            </ComboboxValue>
          </ComboboxChips>
        </ComboboxInputGroup>
        <ComboboxPopup>
          <ComboboxEmpty>No fruit found</ComboboxEmpty>
          <ComboboxList<Fruit>>
            {fruit => (
              <ComboboxItem key={fruit.value} value={fruit}>{fruit.label}</ComboboxItem>
            )}
          </ComboboxList>
        </ComboboxPopup>
      </Combobox>
    </div>
  );
}
