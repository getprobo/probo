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

import { Tabs } from "@probo/ui/src/v2/Tabs/Tabs";
import { TabsIndicator } from "@probo/ui/src/v2/Tabs/TabsIndicator";
import { TabsList } from "@probo/ui/src/v2/Tabs/TabsList";
import { TabsPanel } from "@probo/ui/src/v2/Tabs/TabsPanel";
import { TabsTab } from "@probo/ui/src/v2/Tabs/TabsTab";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskEngagementSection_task$key } from "#/__generated__/core/TaskEngagementSection_task.graphql";

import { taskEngagementSection } from "../variants";

import { TaskActivitySection } from "./TaskActivitySection";
import { TaskCommentsSection } from "./TaskCommentsSection";

const taskEngagementSectionFragment = graphql`
  fragment TaskEngagementSection_task on Task {
    canListComments: permission(action: "core:task-comment:list")
    canListActivities: permission(action: "core:task-activity:list")
    ...TaskCommentsSection_task
    ...TaskActivitySection_task
  }
`;

interface TaskEngagementSectionProps {
  taskKey: TaskEngagementSection_task$key;
}

export function TaskEngagementSection({ taskKey }: TaskEngagementSectionProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskEngagementSectionFragment, taskKey);
  const { root, panel } = taskEngagementSection();
  const defaultTab = task.canListComments ? "comments" : "activity";
  const [tab, setTab] = useState(defaultTab);
  const [activityVisited, setActivityVisited] = useState(
    defaultTab === "activity",
  );

  if (!task.canListComments && !task.canListActivities) {
    return null;
  }

  return (
    <Tabs
      className={root()}
      value={tab}
      onValueChange={(next) => {
        const value = next as string;
        setTab(value);
        if (value === "activity") {
          setActivityVisited(true);
        }
      }}
    >
      <TabsList>
        {task.canListComments && (
          <TabsTab value="comments">
            {t("detailsPage.comments.title")}
          </TabsTab>
        )}
        {task.canListActivities && (
          <TabsTab value="activity">
            {t("detailsPage.activity.title")}
          </TabsTab>
        )}
        <TabsIndicator />
      </TabsList>
      {task.canListComments && (
        <TabsPanel
          value="comments"
          keepMounted
          className={panel()}
        >
          <TaskCommentsSection taskKey={task} />
        </TabsPanel>
      )}
      {task.canListActivities && (
        <TabsPanel
          value="activity"
          keepMounted={activityVisited}
          className={panel()}
        >
          <TaskActivitySection taskKey={task} />
        </TabsPanel>
      )}
    </Tabs>
  );
}
