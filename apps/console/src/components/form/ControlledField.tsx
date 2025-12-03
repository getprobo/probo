import type { ComponentProps } from "react";
import { Field } from "@probo/ui";
import { Controller, type FieldValues } from "react-hook-form";
import { Select } from "@probo/ui";

type Props<T extends typeof Field | typeof Select, TFieldValues extends FieldValues = FieldValues> =
  ComponentProps<T> & Omit<ComponentProps<typeof Controller<TFieldValues>>, "render">;

export function ControlledField<TFieldValues extends FieldValues = FieldValues>({
  control,
  name,
  ...props
}: Props<typeof Field, TFieldValues>) {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <>
          <Field
            {...props}
            {...field}
            // TODO : Find a better way to handle this case (comparing number and string for select create issues)
            value={field.value ? field.value.toString() : ""}
            onValueChange={field.onChange}
          />
        </>
      )}
    />
  );
}

export function ControlledSelect({
  control,
  name,
  ...props
}: Props<typeof Select>) {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          id={name}
          {...props}
          {...field}
          onValueChange={field.onChange}
          value={field.value ?? ""}
        />
      )}
    />
  );
}
