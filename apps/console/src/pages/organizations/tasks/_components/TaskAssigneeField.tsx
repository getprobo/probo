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
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskAssigneeField_task$key } from "#/__generated__/core/TaskAssigneeField_task.graphql";
import { usePeople } from "#/hooks/graph/PeopleGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const noneValue = "__NONE__";

const taskAssigneeFieldFragment = graphql`
  fragment TaskAssigneeField_task on Task {
    assignedTo {
      id
      fullName
    }
  }
`;

interface TaskAssigneeFieldProps {
  taskKey: TaskAssigneeField_task$key;
  disabled?: boolean;
  onValueChange: (value: string | null) => void;
}

export function TaskAssigneeField({
  taskKey,
  disabled,
  onValueChange,
}: TaskAssigneeFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskAssigneeFieldFragment, taskKey);
  const organizationId = useOrganizationId();
  const people = usePeople(organizationId, { contractEnded: false });
  const value = task.assignedTo?.id ?? null;
  const names = new Map(people.map(person => [person.id, person.fullName]));
  const assignedTo = task.assignedTo;
  if (assignedTo) {
    names.set(assignedTo.id, assignedTo.fullName);
  }
  const assignedToMissing = assignedTo != null
    && !people.some(person => person.id === assignedTo.id);

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
        aria-label={t("detailsPage.fields.assignedTo")}
        placeholder={t("detailsPage.unassigned")}
      >
        {(selected: string | null) => {
          if (selected == null || selected === noneValue) {
            return t("detailsPage.unassigned");
          }
          return names.get(selected) ?? selected;
        }}
      </SelectTrigger>
      <SelectPopup>
        <SelectItem value={noneValue}>{t("detailsPage.unassigned")}</SelectItem>
        {assignedToMissing && assignedTo && (
          <SelectItem value={assignedTo.id}>{assignedTo.fullName}</SelectItem>
        )}
        {people.map(person => (
          <SelectItem key={person.id} value={person.id}>
            {person.fullName}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
