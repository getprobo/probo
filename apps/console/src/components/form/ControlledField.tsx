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
