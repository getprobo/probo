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

import { useTranslate } from "@probo/i18n";
import { Avatar, Field, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense } from "react";
import { type Control, Controller, type FieldPath, type FieldValues } from "react-hook-form";

import { usePeople } from "#/hooks/graph/PeopleGraph";

type Props<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
  organizationId: string;
  control: Control<TFieldValues>;
  name: TName;
  label?: string;
  error?: string;
  optional?: boolean;
} & ComponentProps<typeof Field>;

export function PeopleSelectField<TFieldValues extends FieldValues = FieldValues>({
  organizationId,
  control,
  ...props
}: Props<TFieldValues>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" loading placeholder="Loading..." />}
      >
        <PeopleSelectWithQuery<TFieldValues>
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          optional={props.optional}
        />
      </Suspense>
    </Field>
  );
}

function PeopleSelectWithQuery<TFieldValues extends FieldValues = FieldValues>(
  props: Pick<
    Props<TFieldValues>,
    "organizationId" | "control" | "name" | "disabled" | "optional"
  >,
) {
  const { __ } = useTranslate();
  const { name, organizationId, control } = props;
  const people = usePeople(organizationId, { contractEnded: false });

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          disabled={props.disabled}
          id={name}
          variant="editor"
          placeholder={__("Select an owner")}
          onValueChange={value =>
            field.onChange(value === "__NONE__" ? null : value)}
          key={people?.length.toString() ?? "0"}
          {...field}
          className="w-full"
          value={field.value ?? (props.optional ? "__NONE__" : "")}
        >
          {props.optional && <Option value="__NONE__">{__("None")}</Option>}
          {people?.map(p => (
            <Option key={p.id} value={p.id} className="flex gap-2">
              <Avatar name={p.fullName} />
              {p.fullName}
            </Option>
          ))}
        </Select>
      )}
    />
  );
}
