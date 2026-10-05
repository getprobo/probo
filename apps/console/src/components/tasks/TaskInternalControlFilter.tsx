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

import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Suspense, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { TaskInternalControlFilterQuery } from "#/__generated__/core/TaskInternalControlFilterQuery.graphql";
import {
  type InternalControlOption,
  InternalControlSelect,
} from "#/components/form/InternalControlSelect";

const selectedInternalControlQuery = graphql`
  query TaskInternalControlFilterQuery($id: ID!) {
    node(id: $id) @catch(to: RESULT) {
      __typename
      ... on InternalControl {
        id
        name
      }
    }
  }
`;

interface TaskInternalControlFilterProps {
  value: string | null;
  onValueChange: (value: string | null) => void;
}

export function TaskInternalControlFilter({
  value,
  onValueChange,
}: TaskInternalControlFilterProps) {
  const { t } = useTranslation();
  const placeholder = t("tasksCard.filters.allInternalControls");
  const [picked, setPicked] = useState<InternalControlOption | null>(null);
  const known = picked?.id === value ? picked : null;

  const clear = useCallback(() => {
    setPicked(null);
    onValueChange(null);
  }, [onValueChange]);

  function handleValueChange(next: InternalControlOption | null) {
    setPicked(next);
    onValueChange(next?.id ?? null);
  }

  if (value != null && known == null) {
    return (
      <Suspense fallback={<SelectSkeleton size={2} className="w-full" />}>
        <SelectedInternalControlFilter
          id={value}
          placeholder={placeholder}
          searchPlaceholder={t("tasksCard.filters.searchInternalControls")}
          emptyLabel={t("tasksCard.filters.noInternalControls")}
          ariaLabel={t("tasksCard.filters.internalControl")}
          onClear={clear}
          onValueChange={handleValueChange}
        />
      </Suspense>
    );
  }

  return (
    <InternalControlSelect
      value={known}
      allowClear
      size={2}
      placeholder={placeholder}
      searchPlaceholder={t("tasksCard.filters.searchInternalControls")}
      emptyLabel={t("tasksCard.filters.noInternalControls")}
      ariaLabel={t("tasksCard.filters.internalControl")}
      onValueChange={handleValueChange}
    />
  );
}

function SelectedInternalControlFilter({
  id,
  placeholder,
  searchPlaceholder,
  emptyLabel,
  ariaLabel,
  onClear,
  onValueChange,
}: {
  id: string;
  placeholder: string;
  searchPlaceholder: string;
  emptyLabel: string;
  ariaLabel: string;
  onClear: () => void;
  onValueChange: (value: InternalControlOption | null) => void;
}) {
  const data = useLazyLoadQuery<TaskInternalControlFilterQuery>(
    selectedInternalControlQuery,
    { id },
    { fetchPolicy: "network-only" },
  );
  const resolved = data.node.ok && data.node.value.__typename === "InternalControl"
    ? data.node.value
    : null;
  const value = resolved == null ? null : { id: resolved.id, name: resolved.name };

  useEffect(() => {
    if (resolved == null) {
      onClear();
    }
  }, [onClear, resolved]);

  return (
    <InternalControlSelect
      value={value}
      allowClear
      size={2}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyLabel={emptyLabel}
      ariaLabel={ariaLabel}
      onValueChange={onValueChange}
    />
  );
}
