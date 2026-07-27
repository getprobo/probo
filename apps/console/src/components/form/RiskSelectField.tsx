// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Combobox, ComboboxItem, Field, InfiniteScrollTrigger } from "@probo/ui";
import { type ComponentProps, Suspense, useCallback, useMemo, useState } from "react";
import {
  type Control,
  Controller,
  type FieldPath,
  type FieldValues,
} from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useDebounceCallback } from "usehooks-ts";

import { usePaginatedRisks } from "#/hooks/graph/usePaginatedRisks";

type SelectedRisk = {
  id: string;
  name: string;
};

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
  selectedRisk?: SelectedRisk | null;
} & ComponentProps<typeof Field>;

export function RiskSelectField<TFieldValues extends FieldValues = FieldValues>({
  organizationId,
  control,
  disabled,
  optional,
  selectedRisk,
  ...props
}: Props<TFieldValues>) {
  const { t } = useTranslation();

  return (
    <Field {...props}>
      <Suspense
        fallback={(
          <Combobox
            onSearch={() => {}}
            placeholder={t("riskSelectField.placeholder")}
            disabled
          >
            <div />
          </Combobox>
        )}
      >
        <RiskSelectWithQuery<TFieldValues>
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={disabled}
          optional={optional}
          selectedRisk={selectedRisk}
        />
      </Suspense>
    </Field>
  );
}

function RiskSelectWithQuery<TFieldValues extends FieldValues = FieldValues>(
  props: Pick<
    Props<TFieldValues>,
    | "organizationId"
    | "control"
    | "name"
    | "disabled"
    | "optional"
    | "selectedRisk"
  >,
) {
  const { t } = useTranslation();
  const { name, organizationId, control, disabled, optional, selectedRisk }
    = props;
  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginatedRisks(organizationId);
  const [search, setSearch] = useState("");

  const refetchSearch = useDebounceCallback(
    useCallback(
      (query: string) => {
        refetch(
          {
            first: 50,
            filter: { query: query || null },
          },
          { fetchPolicy: "network-only" },
        );
      },
      [refetch],
    ),
    300,
  );

  const handleSearch = (query: string) => {
    setSearch(query);
    refetchSearch(query);
  };

  const risks = useMemo(() => {
    return data?.risks.edges?.map(edge => edge.node) ?? [];
  }, [data?.risks.edges]);

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => {
        const selectedFromList = field.value
          ? risks.find(risk => risk.id === field.value)
          : null;
        const selected
          = selectedFromList
            ?? (selectedRisk && selectedRisk.id === field.value
              ? selectedRisk
              : null);

        return (
          <Combobox
            id={name}
            name={field.name}
            ref={field.ref}
            onBlur={field.onBlur}
            placeholder={t("riskSelectField.placeholder")}
            value={search || selected?.name || ""}
            onSearch={handleSearch}
            disabled={disabled}
          >
            {optional && (
              <ComboboxItem
                onClick={() => {
                  field.onChange(null);
                  setSearch("");
                  refetchSearch("");
                }}
              >
                {t("riskSelectField.none")}
              </ComboboxItem>
            )}
            {risks.map(risk => (
              <ComboboxItem
                key={risk.id}
                onClick={() => {
                  field.onChange(risk.id);
                  setSearch(risk.name);
                }}
              >
                <div className="space-y-1 text-start min-w-0">
                  <div className="max-w-75 ellipsis overflow-hidden whitespace-pre-wrap">
                    {risk.name}
                  </div>
                  {risk.category && (
                    <div className="text-sm text-txt-secondary">
                      {risk.category}
                    </div>
                  )}
                </div>
              </ComboboxItem>
            ))}
            {hasNext && (
              <InfiniteScrollTrigger
                loading={isLoadingNext}
                onView={() => loadNext(50)}
              />
            )}
          </Combobox>
        );
      }}
    />
  );
}
