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

import { Combobox, ComboboxItem, Field } from "@probo/ui";
import { type ComponentProps, Suspense, useMemo, useState } from "react";
import {
  type Control,
  Controller,
  type FieldPath,
  type FieldValues,
} from "react-hook-form";
import { useTranslation } from "react-i18next";

import { usePaginatedInternalControls } from "#/hooks/graph/usePaginatedInternalControls";

type Props<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
  organizationId: string;
  control: Control<TFieldValues>;
  name: TName;
  label?: string;
  error?: string;
  disabled?: boolean;
  optional?: boolean;
} & ComponentProps<typeof Field>;

export function InternalControlSelectField<
  TFieldValues extends FieldValues = FieldValues,
>({
  organizationId,
  control,
  disabled,
  optional,
  ...props
}: Props<TFieldValues>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={(
          <Combobox onSearch={() => {}} placeholder="Loading..." disabled>
            <div />
          </Combobox>
        )}
      >
        <InternalControlSelectWithQuery<TFieldValues>
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={disabled}
          optional={optional}
        />
      </Suspense>
    </Field>
  );
}

function InternalControlSelectWithQuery<TFieldValues extends FieldValues = FieldValues>(
  props: Pick<
    Props<TFieldValues>,
    "organizationId" | "control" | "name" | "disabled" | "optional"
  >,
) {
  const { t } = useTranslation();
  const { name, organizationId, control, disabled, optional } = props;
  const { data } = usePaginatedInternalControls(organizationId);
  const [search, setSearch] = useState("");
  const measures = useMemo(() => {
    return (
      data?.internalControls.edges
        ?.filter(
          edge =>
            edge.node.name.toLowerCase().includes(search.toLowerCase())
            || edge.node.description?.toLowerCase().includes(search.toLowerCase()),
        )
        .map(edge => edge.node) ?? []
    );
  }, [data?.internalControls.edges, search]);

  const allInternalControls = useMemo(() => {
    return data?.internalControls.edges?.map(edge => edge.node) ?? [];
  }, [data?.internalControls.edges]);

  return (
    <div>
      <Controller
        control={control}
        name={name}
        render={({ field }) => {
          const selectedInternalControl = field.value
            ? allInternalControls?.find(m => m.id === field.value)
            : null;

          return (
            <Combobox
              id={name}
              placeholder={t("internalControlSelectField.placeholder")}
              value={selectedInternalControl ? selectedInternalControl.name : search}
              onSearch={setSearch}
              disabled={disabled}
            >
              {optional && (
                <ComboboxItem
                  onClick={() => {
                    field.onChange(null);
                    setSearch("");
                  }}
                >
                  {t("internalControlSelectField.none")}
                </ComboboxItem>
              )}
              {measures?.map(m => (
                <ComboboxItem
                  key={m.id}
                  onClick={() => {
                    field.onChange(m.id);
                    setSearch(m.name);
                  }}
                >
                  <div className="space-y-1 text-start min-w-0">
                    <div className="max-w-75 ellipsis overflow-hidden whitespace-pre-wrap">
                      {m.name}
                    </div>
                    <div className="text-sm text-txt-secondary">
                      {m.category}
                    </div>
                  </div>
                </ComboboxItem>
              ))}
            </Combobox>
          );
        }}
      />
    </div>
  );
}
