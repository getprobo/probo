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

import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskAssigneeField_task$key } from "#/__generated__/core/TaskAssigneeField_task.graphql";
import { UserSelect } from "#/components/form/UserSelect";

const noneValue = "__NONE__";

const taskAssigneeFieldFragment = graphql`
  fragment TaskAssigneeField_task on Task {
    assignedTo {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
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
  const assignedTo = task.assignedTo;
  const emptyLabel = t("detailsPage.unassigned");

  return (
    <UserSelect
      value={assignedTo?.id ?? null}
      disabled={disabled}
      onValueChange={onValueChange}
      emptyLabel={emptyLabel}
      emptyValue={noneValue}
      ariaLabel={t("detailsPage.fields.assignedTo")}
      placeholder={emptyLabel}
      size={1}
      contractEnded={false}
      pinned={assignedTo == null
        ? null
        : {
            id: assignedTo.id,
            fullName: assignedTo.fullName,
            emailAddress: assignedTo.emailAddress,
            avatarUrl: assignedTo.avatar?.downloadUrl,
          }}
    />
  );
}
