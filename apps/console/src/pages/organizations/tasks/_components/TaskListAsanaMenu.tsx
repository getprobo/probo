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

import { Asana } from "@probo/ui/src/Atoms/ThirdParties/Asana";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskListAsanaMenu_task$key } from "#/__generated__/core/TaskListAsanaMenu_task.graphql";
import type { TaskListAsanaMenuUnlinkMutation } from "#/__generated__/core/TaskListAsanaMenuUnlinkMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const taskListAsanaMenuFragment = graphql`
  fragment TaskListAsanaMenu_task on Task {
    id
  }
`;

const unlinkMutation = graphql`
  mutation TaskListAsanaMenuUnlinkMutation($input: UnlinkTaskExternalInput!) {
    unlinkTaskExternal(input: $input) {
      task {
        ...TaskAsanaField_task
        ...TaskListItem_task
      }
    }
  }
`;

export interface TaskListAsanaMenuProps {
  taskKey: TaskListAsanaMenu_task$key;
  linked: boolean;
  onLink?: () => void;
}

export function TaskListAsanaMenu({
  taskKey,
  linked,
  onLink,
}: TaskListAsanaMenuProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskListAsanaMenuFragment, taskKey);
  const [unlink, isUnlinking] = useMutation<TaskListAsanaMenuUnlinkMutation>(
    unlinkMutation,
    {
      successMessage: t("detailsPage.asana.unlinked"),
      errorToast: t("detailsPage.asana.errors.unlink"),
    },
  );

  if (linked) {
    return (
      <DropdownItem
        color="neutral"
        disabled={isUnlinking}
        iconStart={<Asana className="size-4" aria-hidden />}
        onClick={() => {
          void unlink({
            variables: { input: { taskId: task.id } },
          });
        }}
      >
        {t("detailsPage.asana.unlink")}
      </DropdownItem>
    );
  }

  return (
    <DropdownItem
      color="neutral"
      iconStart={<Asana className="size-4" aria-hidden />}
      onClick={onLink}
    >
      {t("detailsPage.asana.linkTo")}
    </DropdownItem>
  );
}
