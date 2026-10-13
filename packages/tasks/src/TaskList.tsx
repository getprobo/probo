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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Collapsible } from "@probo/ui/src/v2/Collapsible/Collapsible";
import { CollapsiblePanel } from "@probo/ui/src/v2/Collapsible/CollapsiblePanel";
import { CollapsibleTrigger } from "@probo/ui/src/v2/Collapsible/CollapsibleTrigger";
import { TooltipProvider } from "@probo/ui/src/v2/Tooltip/TooltipProvider";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { TaskStateIcon } from "./TaskStateIcon";
import { type TaskState, taskStateKeys, taskStates } from "./taskState";
import { taskList } from "./variants";

export interface TaskListItemSource {
  id: string;
  state: TaskState;
}

export interface TaskListProps<T extends TaskListItemSource> {
  items: readonly T[];
  filtered?: boolean;
  renderItem: (item: T, sectionState: TaskState) => ReactNode;
}

export function TaskList<T extends TaskListItemSource>({
  items,
  filtered = false,
  renderItem,
}: TaskListProps<T>) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const slots = taskList();
  const itemsByState = new Map<TaskState, T[]>(
    taskStates.map(state => [state, items.filter(item => item.state === state)]),
  );

  if (items.length === 0) {
    return (
      <Text size={2} color="faint" align="center" className={slots.empty()}>
        {filtered ? t("noResults") : t("empty")}
      </Text>
    );
  }

  return (
    <div className={slots.root()}>
      <TooltipProvider>
        {taskStates
          .filter(state => (itemsByState.get(state)?.length ?? 0) > 0)
          .map((state) => {
            const stateItems = itemsByState.get(state) ?? [];
            return (
              <Collapsible key={state} defaultOpen className={slots.group()}>
                <CollapsibleTrigger className={slots.sectionHeader()}>
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
                  <Badge>{stateItems.length}</Badge>
                </CollapsibleTrigger>
                <CollapsiblePanel>
                  <Card variant="soft" padding="none">
                    <div className={slots.list()}>
                      {stateItems.map(item => renderItem(item, state))}
                    </div>
                  </Card>
                </CollapsiblePanel>
              </Collapsible>
            );
          })}
      </TooltipProvider>
    </div>
  );
}
