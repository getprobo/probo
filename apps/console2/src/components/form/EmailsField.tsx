import { Button, IconPlusLarge, IconTrashCan, Input, Label } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useFieldArray } from "react-hook-form";
import type { Control } from "react-hook-form";
import type { UseFormRegister } from "react-hook-form";

type Props = {
  control: Control<any>;
  register: UseFormRegister<any>;
};

/**
 * A field to handle multiple emails
 */
export function EmailsField({ control, register }: Props) {
  const { __ } = useTranslate();
  const { fields, append, remove } = useFieldArray({
    name: "additionalEmailAddresses",
    control,
  });

  return (
    <fieldset className="space-y-2">
      {fields.length > 0 && <Label>{__("Additional emails")}</Label>}
      {fields.map((field, index) => (
        <div key={field.id} className="flex items-stretch">
          <Input
            className="w-full"
            {...register(`additionalEmailAddresses.${index}`)}
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
        onClick={() => append("")}
      >
        {__("Add email")}
      </Button>
    </fieldset>
  );
}
