import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, IconTrashCan, Input, Label } from "@probo/ui";
import type { ArrayPath, Control, FieldValue, FieldValues, Path, UseFormRegister } from "react-hook-form";
import { useFieldArray } from "react-hook-form";

type Props<TFieldValues extends FieldValues = FieldValues> = {
  control: Control<TFieldValues>;
  register: UseFormRegister<TFieldValues>;
};

/**
 * A field to handle multiple emails
 */
export function EmailsField<
  TFieldValues extends FieldValues = FieldValues,
>({ control, register }: Props<TFieldValues>) {
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
          />
          <Button
            icon={IconTrashCan}
            variant="tertiary"
            onClick={() => remove(index)}
          />
        </div>
      ))}
      <Button
        variant="tertiary"
        type="button"
        icon={IconPlusLarge}
        onClick={() => append("" as FieldValue<TFieldValues>)}
      >
        {__("Add email")}
      </Button>
    </fieldset>
  );
}
