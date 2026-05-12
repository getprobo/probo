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
import { Button, IconPlusLarge, IconTrashCan, Input, Label } from "@probo/ui";
import type { ArrayPath, Control, FieldValue, FieldValues, Path, UseFormRegister } from "react-hook-form";
import { useFieldArray } from "react-hook-form";

type Props<TFieldValues extends FieldValues = FieldValues> = {
  disabled: boolean;
  control: Control<TFieldValues>;
  register: UseFormRegister<TFieldValues>;
};

/**
 * A field to handle multiple emails
 */
export function EmailsField<
  TFieldValues extends FieldValues = FieldValues,
>({ control, register, disabled }: Props<TFieldValues>) {
  const { __ } = useTranslate();
  const { fields, append, remove } = useFieldArray({
    name: "additionalEmailAddresses" as ArrayPath<TFieldValues>,
    control,
  });

  return (
    <fieldset className="space-y-2">
      {fields.length > 0 && <Label>{__("Additional emails")}</Label>}
      {fields.map((field, index) => (
        <div key={field.id} className="flex items-stretch">
          <Input
            className="w-full"
            {...register(`additionalEmailAddresses.${index}` as Path<TFieldValues>)}
            type="email"
            disabled={disabled}
          />
          <Button
            icon={IconTrashCan}
            variant="tertiary"
            onClick={() => remove(index)}
            disabled={disabled}
          />
        </div>
      ))}
      <Button
        variant="tertiary"
        type="button"
        icon={IconPlusLarge}
        onClick={() => append("" as FieldValue<TFieldValues>)}
        disabled={disabled}
      >
        {__("Add email")}
      </Button>
    </fieldset>
  );
}
