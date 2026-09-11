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

import { PlusIcon, XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import type { FocusEvent } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import {
  parseTaskDuration,
  parseTaskDurationAmount,
  stringifyTaskDuration,
  taskDurationsEqual,
  type TaskDurationUnit,
  taskEstimateDurationUnits,
} from "../_lib/taskDuration";
import { useDebouncedSerializedFieldSave } from "../_lib/useSerializedFieldSave";
import { taskDurationField } from "../variants";

const durationFieldAttr = "data-task-duration-field";
const durationPopupAttr = "data-task-duration-popup";
const clearedDuration = "";
const durationSaveDelayMs = 400;

function isFocusInsideDurationField(event: FocusEvent<HTMLElement>) {
  const next = event.relatedTarget;
  if (!(next instanceof Element)) {
    return false;
  }
  const field = event.currentTarget.closest(`[${durationFieldAttr}]`);
  if (field != null && field.contains(next)) {
    return true;
  }
  return next.closest(`[${durationPopupAttr}]`) != null;
}

interface TaskDurationFieldProps {
  value: string | null;
  disabled?: boolean;
  onValueChange: (value: string | null) => void | Promise<unknown>;
  defaultValue?: string;
  addLabel?: string;
  clearLabel?: string;
  amountAriaLabel?: string;
  unitAriaLabel?: string;
  addTitle?: string;
  units?: readonly TaskDurationUnit[];
}

export function TaskDurationField({
  value,
  disabled,
  onValueChange,
  defaultValue = "PT1H",
  addLabel,
  clearLabel,
  amountAriaLabel,
  unitAriaLabel,
  addTitle,
  units = taskEstimateDurationUnits,
}: TaskDurationFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const { root } = taskDurationField();
  const parsed = value ? parseTaskDuration(value) : null;
  const [amountDraft, setAmountDraft] = useState(
    parsed ? String(parsed.amount) : "",
  );
  const [savedValue, setSavedValue] = useState(value);
  const persistDebounced = useDebouncedSerializedFieldSave(async (encoded) => {
    try {
      await onValueChange(encoded === clearedDuration ? null : encoded);
    } catch {
      const failedAmount = encoded === clearedDuration
        ? ""
        : String(parseTaskDuration(encoded)?.amount ?? "");
      const restoredAmount = value
        ? String(parseTaskDuration(value)?.amount ?? "")
        : "";
      setAmountDraft(current => (current === failedAmount ? restoredAmount : current));
    }
  }, durationSaveDelayMs);
  const unit = parsed?.unit ?? "H";
  const availableUnits = units.includes(unit) ? units : [...units, unit];
  const savedAmount = parsed ? String(parsed.amount) : "";

  if (value !== savedValue) {
    const previousAmount = savedValue
      ? String(parseTaskDuration(savedValue)?.amount ?? "")
      : "";
    setSavedValue(value);
    if (amountDraft === previousAmount) {
      const next = value ? parseTaskDuration(value) : null;
      setAmountDraft(next ? String(next.amount) : "");
    }
  }

  function persistIfChanged(next: string | null, immediate = false) {
    if (
      (next != null && value != null && taskDurationsEqual(next, value))
      || next === value
    ) {
      persistDebounced.cancel();
      return;
    }

    const encoded = next ?? clearedDuration;
    persistDebounced.schedule(encoded);
    if (immediate) {
      persistDebounced.flush();
    }
  }

  function encodedAmount(amountText: string, nextUnit: TaskDurationUnit) {
    if (amountText.trim() === "") {
      return undefined;
    }

    const amount = parseTaskDurationAmount(amountText);
    if (amount == null) {
      return undefined;
    }

    return stringifyTaskDuration(amount, nextUnit);
  }

  function saveAmount(amountText: string, nextUnit: TaskDurationUnit, immediate = false) {
    if (amountText.trim() === "") {
      if (!immediate) {
        return false;
      }
      if (!value || parseTaskDuration(value) == null) {
        return false;
      }
      persistIfChanged(null, true);
      return true;
    }

    const next = encodedAmount(amountText, nextUnit);
    if (next == null) {
      return false;
    }

    persistIfChanged(next, immediate);
    return true;
  }

  if (!value) {
    const addButton = (
      <Button
        type="button"
        size={1}
        variant="soft"
        color="neutral"
        iconStart={<PlusIcon />}
        disabled={disabled}
        title={addTitle}
        onClick={() => {
          persistIfChanged(defaultValue, true);
        }}
      >
        {addLabel ?? t("detailsPage.actions.addEstimate")}
      </Button>
    );

    if (disabled && addTitle) {
      return <span title={addTitle}>{addButton}</span>;
    }

    return addButton;
  }

  return (
    <div className={root()} data-task-duration-field="">
      <TextField
        size={1}
        type="number"
        min={1}
        step={1}
        value={amountDraft}
        disabled={disabled}
        aria-label={amountAriaLabel ?? t("detailsPage.fields.timeEstimate")}
        onChange={(event) => {
          setAmountDraft(event.currentTarget.value);
          saveAmount(event.currentTarget.value, unit);
        }}
        onValueChange={(next: string) => {
          setAmountDraft(next);
          saveAmount(next, unit);
        }}
        onBlur={() => {
          if (!saveAmount(amountDraft, unit, true)) {
            setAmountDraft(savedAmount);
          }
        }}
        onKeyDown={(event) => {
          if (event.key !== "Enter") {
            return;
          }
          event.preventDefault();
          if (!saveAmount(amountDraft, unit, true)) {
            setAmountDraft(savedAmount);
          }
        }}
      />
      <Select
        value={unit}
        disabled={disabled}
        onValueChange={(next: TaskDurationUnit | null) => {
          if (next == null) {
            return;
          }
          saveAmount(amountDraft, next, true);
        }}
        onOpenChange={(open, details) => {
          if (open || details.reason === "item-press") {
            return;
          }
          saveAmount(amountDraft, unit, true);
        }}
      >
        <SelectTrigger
          size={1}
          aria-label={unitAriaLabel ?? t("detailsPage.fields.timeEstimateUnit")}
          onBlur={(event) => {
            if (isFocusInsideDurationField(event)) {
              return;
            }
            saveAmount(amountDraft, unit, true);
          }}
        >
          {(selected: TaskDurationUnit | null) =>
            selected
              ? t(`detailsPage.duration.${selected}`)
              : t("detailsPage.duration.H")}
        </SelectTrigger>
        <SelectPopup data-task-duration-popup="">
          {availableUnits.map(item => (
            <SelectItem key={item} value={item}>
              {t(`detailsPage.duration.${item}`)}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
      <IconButton
        type="button"
        size={1}
        variant="ghost"
        color="neutral"
        disabled={disabled}
        aria-label={clearLabel ?? t("detailsPage.actions.clearEstimate")}
        onClick={() => {
          persistIfChanged(null, true);
        }}
      >
        <XIcon aria-hidden />
      </IconButton>
    </div>
  );
}
