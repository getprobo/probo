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
import { type ReactNode, Suspense, useCallback, useEffect, useRef } from "react";
import {
  type Control,
  Controller,
  type FieldPath,
  type FieldValues,
} from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { InternalControlCategoryComboboxQuery } from "#/__generated__/core/InternalControlCategoryComboboxQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const internalControlCategoriesQuery = graphql`
  query InternalControlCategoryComboboxQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        internalControlCategories
      }
    }
  }
`;

type Props<TFieldValues extends FieldValues> = {
  control: Control<TFieldValues>;
  name: FieldPath<TFieldValues>;
};

function categoryToSubmit(categories: readonly string[], value: string): string {
  const trimmed = value.trim();
  const match = categories.find(category =>
    category.toLowerCase() === trimmed.toLowerCase(),
  );
  return match ?? trimmed;
}

export function InternalControlCategoryCombobox<TFieldValues extends FieldValues>(
  props: Props<TFieldValues>,
) {
  const { t } = useTranslation();

  return (
    <Suspense
      fallback={(
        <Input
          id={props.name}
          disabled
          placeholder={t("internalControlFormDialog.fields.categoryPlaceholder")}
        />
      )}
    >
      <InternalControlCategoryComboboxWithQuery {...props} />
    </Suspense>
  );
}

function InternalControlCategoryComboboxWithQuery<TFieldValues extends FieldValues>(
  props: Props<TFieldValues>,
) {
  const { control, name } = props;
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { organization } = useLazyLoadQuery<InternalControlCategoryComboboxQuery>(
    internalControlCategoriesQuery,
    { organizationId },
    { fetchPolicy: "store-and-network" },
  );
  const categories = organization.__typename === "Organization"
    ? organization.internalControlCategories
    : [];

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => {
        const value = typeof field.value === "string" ? field.value : "";
        const trimmed = value.trim();
        const search = trimmed.toLowerCase();
        const matches = categories.filter(category =>
          category.toLowerCase().includes(search),
        );
        const showCreate = search.length > 0 && !categories.some(category =>
          category.toLowerCase() === search,
        );

        return (
          <CategoryField
            categories={categories}
            field={field}
            value={value}
            placeholder={t("internalControlFormDialog.fields.categoryPlaceholder")}
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
            {showCreate && (
              <ComboboxItem
                value={trimmed}
                onClick={() => field.onChange(trimmed)}
              >
                {t("internalControlFormDialog.fields.createCategory", { name: trimmed })}
              </ComboboxItem>
            )}
          </CategoryField>
        );
      }}
    />
  );
}

type CategoryFieldProps = {
  categories: readonly string[];
  field: {
    name: string;
    onBlur: () => void;
    onChange: (value: string) => void;
    ref: (instance: HTMLInputElement | null) => void;
  };
  value: string;
  placeholder: string;
  children: ReactNode;
};

function CategoryField(props: CategoryFieldProps) {
  const { categories, field, value, placeholder, children } = props;
  const { name, onBlur, onChange, ref } = field;
  const inputRef = useRef<HTMLInputElement | null>(null);
  const setInput = useCallback((node: HTMLInputElement | null) => {
    inputRef.current = node;
    ref(node);
  }, [ref]);

  useEffect(() => {
    const form = inputRef.current?.form;
    if (!form) {
      return;
    }

    const normalizeBeforeSubmit = () => {
      const input = inputRef.current;
      if (!input) {
        return;
      }
      const next = categoryToSubmit(categories, input.value);
      if (next !== input.value) {
        onChange(next);
      }
    };

    form.addEventListener("submit", normalizeBeforeSubmit, true);
    return () => form.removeEventListener("submit", normalizeBeforeSubmit, true);
  }, [categories, onChange]);

  return (
    <Combobox
      id={name}
      name={name}
      ref={setInput}
      required
      value={value}
      onSearch={onChange}
      onBlur={() => {
        onBlur();
        const next = categoryToSubmit(categories, value);
        if (next !== value) {
          onChange(next);
        }
      }}
      placeholder={placeholder}
    >
      {children}
    </Combobox>
  );
}
