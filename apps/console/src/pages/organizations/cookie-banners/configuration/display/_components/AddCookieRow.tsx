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

import { toMaxAgeSeconds } from "@probo/helpers";
import { Button, DurationInput, Input, Td, Tr } from "@probo/ui";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

import type { CookieEntry } from "./CategorySection";

interface CookieFormValues {
  name: string;
  duration: { value: string; unit: string };
  description: string;
}

interface AddCookieRowProps {
  isUpdating: boolean;
  onSave: (cookie: CookieEntry) => void;
  onCancel: () => void;
}

export function AddCookieRow({
  isUpdating,
  onSave,
  onCancel,
}: AddCookieRowProps) {
  const { t } = useTranslation("organizations/cookie-banners");

  const { register, handleSubmit, control } = useForm<CookieFormValues>({
    defaultValues: {
      name: "",
      duration: { value: "", unit: "days" },
      description: "",
    },
  });

  const onSubmit = (data: CookieFormValues) => {
    onSave({
      name: data.name,
      maxAgeSeconds: toMaxAgeSeconds(data.duration.value, data.duration.unit),
      description: data.description,
      excluded: false,
    });
  };

  return (
    <Tr>
      <Td className="pr-3">
        <div className="flex flex-col gap-2 min-w-0 max-w-xs">
          <Input
            {...register("name")}
            placeholder={t("addCookieRow.fields.namePlaceholder")}
          />
          <Input
            {...register("description")}
            placeholder={t("addCookieRow.fields.descriptionPlaceholder")}
          />
        </div>
      </Td>
      <Td />
      <Td className="pr-3">
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
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => void handleSubmit(onSubmit)()}
            disabled={isUpdating}
          >
            {t("addCookieRow.actions.save")}
          </Button>
          <Button
            variant="secondary"
            onClick={onCancel}
          >
            {t("addCookieRow.actions.cancel")}
          </Button>
        </div>
      </Td>
    </Tr>
  );
}
