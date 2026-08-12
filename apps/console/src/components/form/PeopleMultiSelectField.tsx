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
import { type ComponentProps, Suspense } from "react";
import { type Control, type FieldValues } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { EntityMultiSelectField } from "#/components/form/EntityMultiSelectField";
import { usePeople } from "#/hooks/graph/PeopleGraph";

type Person = {
  id: string;
  fullName: string;
  emailAddress?: string | null;
};

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: string;
  label?: string;
  error?: string;
  selectedPeople?: Person[];
  placeholder?: string;
} & ComponentProps<typeof Field>;

export function PeopleMultiSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  selectedPeople = [],
  placeholder,
  ...props
}: Props<T>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" disabled placeholder="Loading..." />}
      >
        <PeopleMultiSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          selectedPeople={selectedPeople}
          placeholder={placeholder}
        />
      </Suspense>
    </Field>
  );
}

function PeopleMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<
    Props<T>,
    | "organizationId"
    | "control"
    | "name"
    | "disabled"
    | "selectedPeople"
    | "placeholder"
  >,
) {
  const { t } = useTranslation();
  const {
    name,
    organizationId,
    control,
    selectedPeople = [],
    placeholder,
    disabled,
  } = props;
  const people = usePeople(organizationId, { contractEnded: false });

  const selectedItems = selectedPeople.map(person => ({
    id: person.id,
    fullName: person.fullName,
    emailAddress: person.emailAddress ?? "",
  }));

  return (
    <EntityMultiSelectField
      control={control}
      name={name}
      disabled={disabled}
      items={people}
      selectedItems={selectedItems}
      placeholder={placeholder ?? t("peopleMultiSelectField.addPlaceholder")}
      emptyLabel={t("peopleMultiSelectField.empty")}
      getRemoveAriaLabel={person =>
        t("peopleMultiSelectField.remove", { name: person.fullName })}
      renderOption={person => (
        <div className="flex flex-col">
          <span>{person.fullName}</span>
          {person.emailAddress && (
            <span className="text-xs text-txt-secondary">
              {person.emailAddress}
            </span>
          )}
        </div>
      )}
      renderBadgeLabel={person => <span>{person.fullName}</span>}
    />
  );
}
