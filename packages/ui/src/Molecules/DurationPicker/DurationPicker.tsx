// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import type { HTMLAttributes } from "react";

import { Button } from "../../Atoms/Button/Button";
import { IconPlusLarge } from "../../Atoms/Icons";
import { Input } from "../../Atoms/Input/Input";
import { Option, Select } from "../../Atoms/Select/Select";

type Props = {
  value: string | null;
  onValueChange: (value: string | null) => void;
} & HTMLAttributes<HTMLInputElement>;

const stringify = (value: number | null, unit: string): string | null => {
  if (value === null || value <= 0) return null;

  switch (unit) {
    case "M":
      return `PT${value}M`;
    case "H":
      return `PT${value}H`;
    case "D":
      return `P${value}D`;
    case "W":
      return `P${value * 7}D`;
    default:
      return null;
  }
};

const parse = (value: string): { amount: number; unit: string } | null => {
  const match = value.match(/PT?(\d+)([MDWH])/);
  if (!match) return { amount: 0, unit: "D" };
  const amount = parseInt(match[1], 10) || 0;
  const unit = match[2];
  if (amount % 7 === 0 && unit === "D") {
    return { amount: amount / 7, unit: "W" };
  }
  return { amount, unit };
};

export function DurationPicker({ value, onValueChange, ...props }: Props) {
  const { __ } = useTranslate();
  if (!value) {
    return (
      <div>
        <Button
          variant="secondary"
          icon={IconPlusLarge}
          onClick={() => onValueChange("PT1H")}
        />
      </div>
    );
  }

  const { amount, unit } = parse(value || "") || { amount: 0, unit: "D" };

  return (
    <div className="flex gap-2 w-max">
      <Input
        {...props}
        className="w-25 flex-none"
        type="number"
        step={1}
        value={amount}
        onChange={e =>
          onValueChange(stringify(e.target.valueAsNumber, unit))}
      />
      <Select
        className="w-max flex-none"
        value={unit}
        onValueChange={(v: string) =>
          onValueChange(stringify(amount, v))}
      >
        <Option value="M">{__("Minutes")}</Option>
        <Option value="H">{__("Hours")}</Option>
        <Option value="D">{__("Days")}</Option>
        <Option value="W">{__("Weeks")}</Option>
      </Select>
    </div>
  );
}
