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

import { ChatTextIcon, ClockCounterClockwiseIcon } from "@phosphor-icons/react";
import { Tabs } from "@probo/ui/src/v2/Tabs/Tabs";
import { TabsIndicator } from "@probo/ui/src/v2/Tabs/TabsIndicator";
import { TabsList } from "@probo/ui/src/v2/Tabs/TabsList";
import { TabsPanel } from "@probo/ui/src/v2/Tabs/TabsPanel";
import { TabsTab } from "@probo/ui/src/v2/Tabs/TabsTab";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { taskCommentsEmpty, taskEngagement } from "./variants";

export interface TaskEngagementProps {
  canListComments: boolean;
  canListActivities: boolean;
  comments?: ReactNode;
  activity?: ReactNode;
}

export function TaskEngagement({
  canListComments,
  canListActivities,
  comments,
  activity,
}: TaskEngagementProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { root, panel } = taskEngagement();
  const defaultTab = canListComments ? "comments" : "activity";
  const [tab, setTab] = useState(defaultTab);
  const [activityVisited, setActivityVisited] = useState(defaultTab === "activity");

  if (!canListComments && !canListActivities) {
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
        {canListComments && (
          <TabsTab value="comments">
            {t("comments.title")}
          </TabsTab>
        )}
        {canListActivities && (
          <TabsTab value="activity">
            {t("activity.title")}
          </TabsTab>
        )}
        <TabsIndicator />
      </TabsList>
      {canListComments && (
        <TabsPanel value="comments" keepMounted className={panel()}>
          {comments}
        </TabsPanel>
      )}
      {canListActivities && (
        <TabsPanel
          value="activity"
          keepMounted={activityVisited}
          className={panel()}
        >
          {activityVisited ? activity : null}
        </TabsPanel>
      )}
    </Tabs>
  );
}

export function TaskCommentsEmpty() {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { root, icon } = taskCommentsEmpty();

  return (
    <div className={root()}>
      <span className={icon()}>
        <ChatTextIcon weight="light" />
      </span>
      <Text size={1} color="faint">
        {t("comments.empty")}
      </Text>
    </div>
  );
}

export function TaskActivityEmpty() {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { root, icon } = taskCommentsEmpty();

  return (
    <div className={root()}>
      <span className={icon()}>
        <ClockCounterClockwiseIcon weight="light" />
      </span>
      <Text size={1} color="faint">
        {t("activity.empty")}
      </Text>
    </div>
  );
}
