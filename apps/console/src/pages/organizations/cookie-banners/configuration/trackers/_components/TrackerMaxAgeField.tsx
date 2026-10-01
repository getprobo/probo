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
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { type FocusEvent } from "react";
import { useTranslation } from "react-i18next";

import { trackerMaxAgeField } from "../../../variants";

interface TrackerMaxAgeFieldProps {
  id?: string;
  value: string;
  unit: string;
  disabled?: boolean;
  onValueChange: (value: string) => void;
  onUnitChange: (unit: string) => void;
  onBlur?: (event: FocusEvent<HTMLDivElement>) => void;
}

export function TrackerMaxAgeField({
  id,
  value,
  unit,
  disabled,
  onValueChange,
  onUnitChange,
  onBlur,
}: TrackerMaxAgeFieldProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { root, value: valueSlot, unit: unitSlot } = trackerMaxAgeField();

  return (
    <div className={root()} onBlur={onBlur}>
      <TextField
        id={id}
        size={2}
        type="number"
        min={0}
        value={value}
        disabled={disabled}
        placeholder="—"
        className={valueSlot()}
        aria-label={t("trackerProperties.properties.maxAge")}
        onValueChange={onValueChange}
      />
      <div className={unitSlot()}>
        <Select
          value={unit}
          disabled={disabled}
          onValueChange={(next: string | null) => {
            if (next != null) {
              onUnitChange(next);
            }
          }}
        >
          <SelectTrigger
            size={2}
            variant="surface"
            aria-label={t("trackerProperties.properties.maxAgeUnit")}
          >
            {(selected: string | null) =>
              DURATION_UNITS.find(durationUnit => durationUnit.value === selected)?.label
              ?? selected}
          </SelectTrigger>
          <SelectPopup align="end">
            {DURATION_UNITS.map(durationUnit => (
              <SelectItem key={durationUnit.value} value={durationUnit.value}>
                {durationUnit.label}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
    </div>
  );
}
