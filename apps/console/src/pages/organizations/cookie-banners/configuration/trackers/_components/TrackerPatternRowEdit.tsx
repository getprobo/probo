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

import { fromMaxAgeSeconds, toMaxAgeSeconds } from "@probo/helpers";
import { Button, DurationInput, Input, Td, Tr } from "@probo/ui";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

interface FormValues {
  duration: { value: string; unit: string };
  description: string;
}

interface TrackerPatternRowEditProps {
  pattern: string;
  description: string;
  maxAgeSeconds: number | null;
  isUpdating: boolean;
  onSave: (data: { description: string; maxAgeSeconds: number | null }) => void;
  onCancel: () => void;
}

export function TrackerPatternRowEdit({
  pattern,
  description,
  maxAgeSeconds,
  isUpdating,
  onSave,
  onCancel,
}: TrackerPatternRowEditProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const initial = fromMaxAgeSeconds(maxAgeSeconds);

  const { register, handleSubmit, control } = useForm<FormValues>({
    defaultValues: {
      duration: initial,
      description,
    },
  });

  const onSubmit = (data: FormValues) => {
    onSave({
      description: data.description,
      maxAgeSeconds: toMaxAgeSeconds(data.duration.value, data.duration.unit),
    });
  };

  return (
    <Tr>
      <Td colSpan={7}>
        <div className="flex flex-col gap-3">
          <span className="font-medium wrap-break-word">{pattern}</span>
          <div className="flex items-end gap-2">
            <div className="flex flex-col gap-1 flex-1">
              <label className="text-xs text-txt-tertiary">{t("trackerPatternRowEdit.fields.description")}</label>
              <Input
                {...register("description")}
                placeholder={t("trackerPatternRowEdit.fields.descriptionPlaceholder")}
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-txt-tertiary">{t("trackerPatternRowEdit.fields.maxAge")}</label>
              <Controller
                name="duration"
                control={control}
                render={({ field }) => (
                  <DurationInput
                    value={field.value.value}
                    unit={field.value.unit}
                    onValueChange={v => field.onChange({ ...field.value, value: v })}
                    onUnitChange={u => field.onChange({ ...field.value, unit: u })}
                  />
                )}
              />
            </div>
            <Button
              onClick={() => void handleSubmit(onSubmit)()}
              disabled={isUpdating}
            >
              {t("trackerPatternRowEdit.actions.save")}
            </Button>
            <Button
              variant="secondary"
              onClick={onCancel}
            >
              {t("trackerPatternRowEdit.actions.cancel")}
            </Button>
          </div>
        </div>
      </Td>
    </Tr>
  );
}
