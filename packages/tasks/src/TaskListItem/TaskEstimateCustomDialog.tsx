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

import { Form } from "@base-ui/react/form";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import type { ReactNode } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import {
  parseTaskDuration,
  parseTaskDurationAmount,
  stringifyTaskDuration,
  taskDurationsEqual,
  type TaskDurationUnit,
  taskEstimateDurationUnits,
} from "../taskDuration";

const estimateUnits: readonly TaskDurationUnit[] = taskEstimateDurationUnits;

export function TaskEstimateCustomDialog({
  open,
  timeEstimate,
  onOpenChange,
  onSave,
}: {
  open: boolean;
  timeEstimate?: string | null;
  onOpenChange: (open: boolean) => void;
  onSave: (timeEstimate: string) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [amount, setAmount] = useState("1");
  const [unit, setUnit] = useState<TaskDurationUnit>("H");
  const [wasOpen, setWasOpen] = useState(false);
  const parsedAmount = parseTaskDurationAmount(amount);
  const nextEstimate = parsedAmount == null
    ? null
    : stringifyTaskDuration(parsedAmount, unit);
  const availableUnits = estimateUnits.includes(unit)
    ? estimateUnits
    : [unit, ...estimateUnits];

  if (open !== wasOpen) {
    setWasOpen(open);
    if (open) {
      const parsed = timeEstimate ? parseTaskDuration(timeEstimate) : null;
      setAmount(parsed ? String(Math.max(1, Math.round(parsed.amount))) : "1");
      setUnit(parsed?.unit ?? "H");
    }
  }

  function save() {
    if (nextEstimate == null) {
      return;
    }
    if (timeEstimate != null && taskDurationsEqual(nextEstimate, timeEstimate)) {
      onOpenChange(false);
      return;
    }
    onSave(nextEstimate);
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("fields.timeEstimate")}</DialogTitle>
        </DialogHeader>
        <Form className="flex flex-col gap-4" onFormSubmit={save}>
          <DialogBody className="flex flex-col gap-3">
            <Field label={t("fields.timeEstimate")} required>
              <TextField
                type="number"
                min={1}
                step={1}
                required
                value={amount}
                onValueChange={setAmount}
              />
            </Field>
            <Field label={t("fields.timeEstimateUnit")}>
              <Select<TaskDurationUnit>
                value={unit}
                onValueChange={(next) => {
                  if (next == null) {
                    return;
                  }
                  setUnit(next);
                }}
              >
                <SelectTrigger aria-label={t("fields.timeEstimateUnit")}>
                  {(selected: TaskDurationUnit | null) =>
                    selected ? t(`units.${selected}`) : t("units.H")}
                </SelectTrigger>
                <SelectPopup>
                  {availableUnits.map(item => (
                    <SelectItem key={item} value={item}>
                      {t(`units.${item}`)}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </Field>
          </DialogBody>
          <DialogFooter>
            <DialogClose
              render={(
                <Button type="button" variant="soft" color="neutral">
                  {t("durationPicker.cancel")}
                </Button>
              )}
            />
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              disabled={nextEstimate == null}
            >
              {t("durationPicker.save")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}
