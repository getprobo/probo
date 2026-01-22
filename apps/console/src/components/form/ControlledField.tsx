import { Field, Select } from "@probo/ui";
import type { ComponentProps } from "react";
import { Controller, type FieldPath, type FieldValues } from "react-hook-form";

type Props<
  T extends typeof Field | typeof Select,
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>
  = ComponentProps<T> & Omit<ComponentProps<typeof Controller<TFieldValues, TName>>, "render">;

export function ControlledField<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
  control,
  name,
  ...props
}: Props<typeof Field, TFieldValues, TName>) {
  return (
    <Controller<TFieldValues, TName>
      control={control}
      name={name}
      render={({ field }) => (
        <>
          <Field
            {...props}
            {...field}
            // TODO : Find a better way to handle this case (comparing number and string for select create issues)
            value={field.value ? (field.value as readonly string[] | string | number).toString() : ""}
            onValueChange={field.onChange}
          />
        </>
      )}
    />
  );
}

export function ControlledSelect<TFieldValues extends FieldValues = FieldValues>({
  control,
  name,
  ...props
}: Props<typeof Select, TFieldValues>) {
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
