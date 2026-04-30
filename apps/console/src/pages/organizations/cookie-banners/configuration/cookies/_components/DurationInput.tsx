// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { Input } from "@probo/ui";

const UNITS: { value: string; label: string; seconds: number }[] = [
  { value: "seconds", label: "seconds", seconds: 1 },
  { value: "minutes", label: "minutes", seconds: 60 },
  { value: "hours", label: "hours", seconds: 3600 },
  { value: "days", label: "days", seconds: 86400 },
  { value: "weeks", label: "weeks", seconds: 604800 },
  { value: "months", label: "months", seconds: 2592000 },
  { value: "years", label: "years", seconds: 31536000 },
];

export function toMaxAgeSeconds(value: string, unit: string): number | null {
  const num = parseFloat(value);
  if (isNaN(num) || num <= 0) return null;
  const u = UNITS.find(u => u.value === unit);
  if (!u) return null;
  return Math.round(num * u.seconds);
}

export function fromMaxAgeSeconds(seconds: number | null): { value: string; unit: string } {
  if (seconds === null || seconds <= 0) return { value: "", unit: "days" };
  for (const u of [...UNITS].reverse()) {
    if (seconds >= u.seconds && seconds % u.seconds === 0) {
      return { value: String(seconds / u.seconds), unit: u.value };
    }
  }
  return { value: String(seconds), unit: "seconds" };
}

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
        {UNITS.map(u => (
          <option key={u.value} value={u.value}>{u.label}</option>
        ))}
      </select>
    </div>
  );
}
