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

import { DURATION_UNITS } from "@probo/helpers";

import { Input } from "./Input";

interface DurationInputProps {
  value: string;
  unit: string;
  onValueChange: (value: string) => void;
  onUnitChange: (unit: string) => void;
}

export function DurationInput({ value, unit, onValueChange, onUnitChange }: DurationInputProps) {
  return (
    <div className="flex gap-1">
      <Input
        type="number"
        min={0}
        value={value}
        onChange={e => onValueChange(e.target.value)}
        placeholder="—"
        className="w-20"
      />
      <select
        value={unit}
        onChange={e => onUnitChange(e.target.value)}
        className="rounded border border-border bg-background px-2 py-1 text-sm"
      >
        {DURATION_UNITS.map(u => (
          <option key={u.value} value={u.value}>{u.label}</option>
        ))}
      </select>
    </div>
  );
}
