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

import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import { type TaskState, taskStateKeys, taskStates } from "../taskState";
import { TaskStateIcon } from "../TaskStateIcon";
import { taskListItem } from "../variants";

function stateOption(
  state: TaskState,
  slots: ReturnType<typeof taskListItem>,
  t: (key: string) => string,
): ReactNode {
  return (
    <span className={slots.stateOption()}>
      <TaskStateIcon state={state} />
      {t(`states.${taskStateKeys[state]}`)}
    </span>
  );
}

export function TaskStateMenu({
  displayState,
  taskState,
  slots,
  onChange,
}: {
  displayState: TaskState;
  taskState: TaskState;
  slots: ReturnType<typeof taskListItem>;
  onChange: (state: TaskState) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);

  return (
    <DropdownRadioGroup
      value={displayState}
      onValueChange={(state) => {
        if (typeof state !== "string" || state === taskState) {
          return;
        }
        onChange(state as TaskState);
      }}
    >
      {taskStates.map(state => (
        <DropdownRadioItem key={state} value={state}>
          {stateOption(state, slots, t)}
        </DropdownRadioItem>
      ))}
    </DropdownRadioGroup>
  );
}
