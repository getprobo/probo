// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { Button, IconPlusLarge, IconTrashCan, Input, Label } from "@probo/ui";
import type {
  ArrayPath,
  Control,
  FieldValue,
  FieldValues,
  Path,
  UseFormRegister,
} from "react-hook-form";
import { useFieldArray } from "react-hook-form";
import { useTranslation } from "react-i18next";

type Props<TFieldValues extends FieldValues = FieldValues> = {
  disabled: boolean;
  control: Control<TFieldValues>;
  register: UseFormRegister<TFieldValues>;
};

/**
 * A field to handle multiple emails
 */
export function EmailsField<TFieldValues extends FieldValues = FieldValues>({
  control,
  register,
  disabled,
}: Props<TFieldValues>) {
  const { t } = useTranslation();
  const { fields, append, remove } = useFieldArray({
    name: "additionalEmailAddresses" as ArrayPath<TFieldValues>,
    control,
  });

  return (
    <fieldset className="space-y-2">
      {fields.length > 0 && <Label>{t("emailsField.additionalEmails")}</Label>}
      {fields.map((field, index) => (
        <div key={field.id} className="flex items-stretch">
          <Input
            className="w-full"
            {...register(
              `additionalEmailAddresses.${index}` as Path<TFieldValues>,
            )}
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
        {t("emailsField.actions.add")}
      </Button>
    </fieldset>
  );
}
