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

import { toMaxAgeSeconds } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, DurationInput, Input, Td, Tr } from "@probo/ui";
import { Controller, useForm } from "react-hook-form";

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
  const { __ } = useTranslate();

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
        <Input
          {...register("name")}
          placeholder={__("Cookie name")}
        />
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
      <Td className="pr-3">
        <Input
          {...register("description")}
          placeholder={__("Description")}
        />
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => void handleSubmit(onSubmit)()}
            disabled={isUpdating}
          >
            {__("Save")}
          </Button>
          <Button
            variant="secondary"
            onClick={onCancel}
          >
            {__("Cancel")}
          </Button>
        </div>
      </Td>
    </Tr>
  );
}
