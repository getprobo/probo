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
