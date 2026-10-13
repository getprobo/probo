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

import { CaretDownIcon, CaretRightIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { type TaskState, taskStateKeys } from "./taskState";
import { TaskStateIcon } from "./TaskStateIcon";
import { taskList } from "./variants";

export interface TaskListGroupHeaderProps {
  state: TaskState;
  count: number;
  open: boolean;
  offset?: boolean;
  onOpenChange: (open: boolean) => void;
}

export function TaskListGroupHeader({
  state,
  count,
  open,
  offset = false,
  onOpenChange,
}: TaskListGroupHeaderProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const slots = taskList({ open, offset });

  return (
    <button
      type="button"
      className={slots.sectionHeader()}
      aria-expanded={open}
      onClick={() => {
        onOpenChange(!open);
      }}
    >
      <CaretRightIcon
        className={slots.caret()}
        data-slot="caret-closed"
        aria-hidden
      />
      <CaretDownIcon
        className={slots.caret()}
        data-slot="caret-open"
        aria-hidden
      />
      <TaskStateIcon state={state} />
      <Text size={2} weight="medium">
        {t(`states.${taskStateKeys[state]}`)}
      </Text>
      <Badge>{count}</Badge>
    </button>
  );
}
