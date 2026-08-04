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

import { Combobox, ComboboxItem, InfiniteScrollTrigger } from "@probo/ui";
import { Suspense, useState } from "react";
import {
  type Control,
  Controller,
  type FieldPath,
  type FieldValues,
} from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { AuditProgramSelectFieldSelectedQuery } from "#/__generated__/core/AuditProgramSelectFieldSelectedQuery.graphql";
import { usePaginatedAuditPrograms } from "#/hooks/graph/usePaginatedAuditPrograms";

type SelectedAuditProgram = {
  id: string;
  name: string;
  framework?: { id: string } | null;
};

type Props<T extends FieldValues> = {
  organizationId: string;
  frameworkId?: string | null;
  control: Control<T>;
  name: FieldPath<T>;
  disabled?: boolean;
  selectedAuditProgram?: SelectedAuditProgram | null;
};

const selectedAuditProgramQuery = graphql`
  query AuditProgramSelectFieldSelectedQuery($id: ID!) {
    node(id: $id) {
      ... on AuditProgram {
        id
        name
        framework {
          id
        }
      }
    }
  }
`;

function matchesFramework(
  program: SelectedAuditProgram | null | undefined,
  frameworkId?: string | null,
): boolean {
  if (!program) {
    return false;
  }
  if (!frameworkId) {
    return true;
  }
  return !program.framework?.id || program.framework.id === frameworkId;
}

export function AuditProgramSelectField<T extends FieldValues>(props: Props<T>) {
  // Remount when framework changes so search state resets without an effect.
  return (
    <AuditProgramSelectFieldInner
      key={props.frameworkId ?? ""}
      {...props}
    />
  );
}

function AuditProgramSelectFieldInner<T extends FieldValues>({
  organizationId,
  frameworkId,
  control,
  name,
  disabled,
  selectedAuditProgram,
}: Props<T>) {
  const { t } = useTranslation("organizations/audit-programs");
  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginatedAuditPrograms(organizationId);
  const [search, setSearch] = useState("");

  const allPrograms
    = data?.auditPrograms?.edges
      ?.map(edge => edge.node)
      .filter((node): node is NonNullable<typeof node> => node !== null) ?? [];

  const programs = allPrograms.filter(program =>
    matchesFramework(program, frameworkId),
  );

  const filteredPrograms = programs.filter(program =>
    program.name.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => {
        const selectedFromList = field.value
          ? allPrograms.find(program => program.id === field.value)
          : null;
        const selectedFromProp
          = selectedAuditProgram && selectedAuditProgram.id === field.value
            ? selectedAuditProgram
            : null;
        const selectedCandidate = selectedFromList ?? selectedFromProp ?? null;
        const selected = matchesFramework(selectedCandidate, frameworkId)
          ? selectedCandidate
          : null;
        const needsFetch = Boolean(field.value) && !selectedCandidate;

        return (
          <Suspense
            fallback={(
              <AuditProgramCombobox
                fieldName={String(field.name)}
                selected={null}
                search={search}
                onSearch={setSearch}
                onChange={field.onChange}
                onBlur={field.onBlur}
                programs={filteredPrograms}
                disabled={disabled}
                hasNext={hasNext}
                isLoadingNext={isLoadingNext}
                loadNext={loadNext}
                noneLabel={t("auditProgramSelect.none")}
              />
            )}
          >
            {needsFetch
              ? (
                  <AuditProgramComboboxWithFetchedSelected
                    fieldName={String(field.name)}
                    value={field.value}
                    search={search}
                    onSearch={setSearch}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                    programs={filteredPrograms}
                    frameworkId={frameworkId}
                    disabled={disabled}
                    hasNext={hasNext}
                    isLoadingNext={isLoadingNext}
                    loadNext={loadNext}
                    noneLabel={t("auditProgramSelect.none")}
                  />
                )
              : (
                  <AuditProgramCombobox
                    fieldName={String(field.name)}
                    selected={selected}
                    search={search}
                    onSearch={setSearch}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                    programs={filteredPrograms}
                    disabled={disabled}
                    hasNext={hasNext}
                    isLoadingNext={isLoadingNext}
                    loadNext={loadNext}
                    noneLabel={t("auditProgramSelect.none")}
                  />
                )}
          </Suspense>
        );
      }}
    />
  );
}

function AuditProgramComboboxWithFetchedSelected({
  value,
  frameworkId,
  ...props
}: {
  fieldName: string;
  value: string;
  search: string;
  onSearch: (query: string) => void;
  onChange: (value: string) => void;
  onBlur: () => void;
  programs: SelectedAuditProgram[];
  frameworkId?: string | null;
  disabled?: boolean;
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: (count: number) => void;
  noneLabel: string;
}) {
  const data = useLazyLoadQuery<AuditProgramSelectFieldSelectedQuery>(
    selectedAuditProgramQuery,
    { id: value },
    { fetchPolicy: "store-or-network" },
  );

  const fetched
    = data.node?.id
      ? {
          id: data.node.id,
          name: data.node.name ?? "",
          framework: data.node.framework,
        }
      : null;
  const selected = matchesFramework(fetched, frameworkId) ? fetched : null;

  return (
    <AuditProgramCombobox
      {...props}
      selected={selected}
    />
  );
}

function AuditProgramCombobox({
  fieldName,
  selected,
  search,
  onSearch,
  onChange,
  onBlur,
  programs,
  disabled,
  hasNext,
  isLoadingNext,
  loadNext,
  noneLabel,
}: {
  fieldName: string;
  selected: SelectedAuditProgram | null;
  search: string;
  onSearch: (query: string) => void;
  onChange: (value: string) => void;
  onBlur: () => void;
  programs: SelectedAuditProgram[];
  disabled?: boolean;
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: (count: number) => void;
  noneLabel: string;
}) {
  return (
    <Combobox
      id={fieldName}
      name={fieldName}
      onBlur={onBlur}
      placeholder={noneLabel}
      value={search || selected?.name || ""}
      onSearch={onSearch}
      disabled={disabled}
    >
      <ComboboxItem
        onClick={() => {
          onChange("");
          onSearch("");
        }}
      >
        {noneLabel}
      </ComboboxItem>
      {programs.map(program => (
        <ComboboxItem
          key={program.id}
          onClick={() => {
            onChange(program.id);
            onSearch(program.name);
          }}
        >
          {program.name}
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
}
