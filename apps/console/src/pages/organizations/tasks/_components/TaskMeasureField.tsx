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

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskMeasureField_task$key } from "#/__generated__/core/TaskMeasureField_task.graphql";
import { usePaginatedMeasures } from "#/hooks/graph/usePaginatedMeasures";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const noneValue = "__NONE__";
const measurePageSize = 20;
const loadMoreScrollThresholdPx = 48;

const taskMeasureFieldFragment = graphql`
  fragment TaskMeasureField_task on Task {
    measure {
      id
      name
    }
  }
`;

interface TaskMeasureFieldProps {
  taskKey: TaskMeasureField_task$key;
  disabled?: boolean;
  onValueChange: (value: string | null) => void;
}

export function TaskMeasureField({
  taskKey,
  disabled,
  onValueChange,
}: TaskMeasureFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskMeasureFieldFragment, taskKey);
  const organizationId = useOrganizationId();
  const { data, hasNext, isLoadingNext, loadNext } = usePaginatedMeasures(
    organizationId,
    {
      first: measurePageSize,
      order: { field: "NAME", direction: "ASC" },
    },
  );

  const measures = useMemo(
    () => data?.measures.edges.map(edge => edge.node) ?? [],
    [data?.measures.edges],
  );
  const value = task.measure?.id ?? null;
  const names = new Map(measures.map(measure => [measure.id, measure.name]));
  const linkedMeasure = task.measure;
  if (linkedMeasure) {
    names.set(linkedMeasure.id, linkedMeasure.name);
  }
  const linkedMeasureMissing = linkedMeasure != null
    && !measures.some(measure => measure.id === linkedMeasure.id);

  function loadMore() {
    if (hasNext && !isLoadingNext) {
      loadNext(measurePageSize);
    }
  }

  return (
    <Select
      value={value ?? noneValue}
      disabled={disabled}
      onValueChange={(next: string | null) => {
        if (next == null || next === value || (next === noneValue && value == null)) {
          return;
        }
        onValueChange(next === noneValue ? null : next);
      }}
    >
      <SelectTrigger
        size={1}
        aria-label={t("detailsPage.fields.measure")}
        placeholder={t("detailsPage.none")}
      >
        {(selected: string | null) => {
          if (selected == null || selected === noneValue) {
            return t("detailsPage.none");
          }
          return names.get(selected) ?? selected;
        }}
      </SelectTrigger>
      <SelectPopup
        onScroll={(event) => {
          const popup = event.currentTarget;
          const remaining = popup.scrollHeight - popup.scrollTop - popup.clientHeight;
          if (remaining <= loadMoreScrollThresholdPx) {
            loadMore();
          }
        }}
      >
        <SelectItem value={noneValue}>{t("detailsPage.none")}</SelectItem>
        {linkedMeasureMissing && linkedMeasure && (
          <SelectItem value={linkedMeasure.id}>{linkedMeasure.name}</SelectItem>
        )}
        {measures.map(measure => (
          <SelectItem key={measure.id} value={measure.id}>
            {measure.name}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
