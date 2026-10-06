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

import type { TaskInternalControlField_task$key } from "#/__generated__/core/TaskInternalControlField_task.graphql";
import { usePaginatedInternalControls } from "#/hooks/graph/usePaginatedInternalControls";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const noneValue = "__NONE__";
const internalControlPageSize = 20;
const loadMoreScrollThresholdPx = 48;

const taskInternalControlFieldFragment = graphql`
  fragment TaskInternalControlField_task on Task {
    internalControl {
      id
      name
    }
  }
`;

interface TaskInternalControlFieldProps {
  taskKey: TaskInternalControlField_task$key;
  disabled?: boolean;
  onValueChange: (value: string | null) => void;
}

export function TaskInternalControlField({
  taskKey,
  disabled,
  onValueChange,
}: TaskInternalControlFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskInternalControlFieldFragment, taskKey);
  const organizationId = useOrganizationId();
  const { data, hasNext, isLoadingNext, loadNext } = usePaginatedInternalControls(
    organizationId,
    {
      first: internalControlPageSize,
      order: { field: "NAME", direction: "ASC" },
    },
  );

  const internalControls = useMemo(
    () => data?.internalControls.edges.map(edge => edge.node) ?? [],
    [data?.internalControls.edges],
  );
  const value = task.internalControl?.id ?? null;
  const names = new Map(internalControls.map(internalControl => [internalControl.id, internalControl.name]));
  const linkedInternalControl = task.internalControl;
  if (linkedInternalControl) {
    names.set(linkedInternalControl.id, linkedInternalControl.name);
  }
  const linkedInternalControlMissing = linkedInternalControl != null
    && !internalControls.some(internalControl => internalControl.id === linkedInternalControl.id);

  function loadMore() {
    if (hasNext && !isLoadingNext) {
      loadNext(internalControlPageSize);
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
        aria-label={t("detailsPage.fields.internalControl")}
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
        {linkedInternalControlMissing && linkedInternalControl && (
          <SelectItem value={linkedInternalControl.id}>{linkedInternalControl.name}</SelectItem>
        )}
        {internalControls.map(internalControl => (
          <SelectItem key={internalControl.id} value={internalControl.id}>
            {internalControl.name}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
