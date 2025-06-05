import type { ComponentProps, JSX, JSXElementConstructor } from "react";
import { Field } from "@probo/ui";
import { Controller, type Control } from "react-hook-form";
import { Select } from "@probo/ui";

type Props<T extends keyof JSX.IntrinsicElements | JSXElementConstructor<any>> =
  ComponentProps<T> & {
    control: Control<any>;
    name: string;
  };

export function ControlledField({
  control,
  name,
  ...props
}: Props<typeof Field>) {
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
