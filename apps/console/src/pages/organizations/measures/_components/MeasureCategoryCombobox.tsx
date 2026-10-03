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

import { Combobox, ComboboxItem, Input } from "@probo/ui";
import { Suspense } from "react";
import {
  type Control,
  Controller,
  type FieldPath,
  type FieldValues,
} from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { MeasureCategoryComboboxQuery } from "#/__generated__/core/MeasureCategoryComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const measureCategoriesQuery = graphql`
  query MeasureCategoryComboboxQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        measureCategories
      }
    }
  }
`;

type Props<TFieldValues extends FieldValues> = {
  control: Control<TFieldValues>;
  name: FieldPath<TFieldValues>;
};

export function MeasureCategoryCombobox<TFieldValues extends FieldValues>(
  props: Props<TFieldValues>,
) {
  const { t } = useTranslation();

  return (
    <Suspense
      fallback={(
        <Input
          disabled
          placeholder={t("measureFormDialog.fields.categoryPlaceholder")}
        />
      )}
    >
      <MeasureCategoryComboboxWithQuery {...props} />
    </Suspense>
  );
}

function MeasureCategoryComboboxWithQuery<TFieldValues extends FieldValues>(
  props: Props<TFieldValues>,
) {
  const { control, name } = props;
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { organization } = useLazyLoadQuery<MeasureCategoryComboboxQuery>(
    measureCategoriesQuery,
    { organizationId },
    { fetchPolicy: "store-and-network" },
  );
  const categories = organization.__typename === "Organization"
    ? organization.measureCategories
    : [];

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => {
        const value = (field.value as string | undefined) ?? "";
        const trimmed = value.trim();
        const search = trimmed.toLowerCase();
        const matches = categories.filter(c => c.toLowerCase().includes(search));
        const exists = categories.some(c => c.toLowerCase() === search);

        return (
          <Combobox
            id={name}
            required
            value={value}
            onSearch={field.onChange}
            onBlur={field.onBlur}
            placeholder={t("measureFormDialog.fields.categoryPlaceholder")}
          >
            {matches.map(category => (
              <ComboboxItem
                key={category}
                value={category}
                onClick={() => field.onChange(category)}
              >
                {category}
              </ComboboxItem>
            ))}
            {search && !exists && (
              <ComboboxItem
                value={trimmed}
                onClick={() => field.onChange(trimmed)}
              >
                {t("measureFormDialog.fields.createCategory", { name: trimmed })}
              </ComboboxItem>
            )}
          </Combobox>
        );
      }}
    />
  );
}
