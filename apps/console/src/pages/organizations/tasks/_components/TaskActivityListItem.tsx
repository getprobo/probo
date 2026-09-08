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

import { UserIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskActivityListItem_activity$key } from "#/__generated__/core/TaskActivityListItem_activity.graphql";

import {
  formatTaskActivityValue,
  taskActivitySentenceKey,
} from "../_lib/taskActivity";
import { taskCommentListItem } from "../variants";

const taskActivityListItemFragment = graphql`
  fragment TaskActivityListItem_activity on TaskActivity {
    activityType
    field
    oldValue
    newValue
    createdAt
    actor {
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
  }
`;

interface TaskActivityListItemProps {
  taskActivityKey: TaskActivityListItem_activity$key;
}

export function TaskActivityListItem({ taskActivityKey }: TaskActivityListItemProps) {
  const { i18n, t } = useTranslation("organizations/tasks");
  const activity = useFragment(taskActivityListItemFragment, taskActivityKey);
  const { header } = taskCommentListItem();
  const actorName = activity.actor?.fullName
    ?? t("detailsPage.activity.actorFallback");
  const field = activity.field;
  const oldValue = formatTaskActivityValue(
    field,
    activity.oldValue,
    i18n.language,
    t,
  );
  const newValue = formatTaskActivityValue(
    field,
    activity.newValue,
    i18n.language,
    t,
  );
  const sentence = t(
    taskActivitySentenceKey(
      activity.activityType,
      field,
      oldValue,
      newValue,
    ),
    {
      actor: actorName,
      oldValue,
      newValue,
    },
  );

  return (
    <ListItem>
      <Avatar
        name={activity.actor?.fullName}
        email={activity.actor?.emailAddress}
        src={activity.actor?.avatar?.downloadUrl}
        size={2}
        radius="full"
        fallback={activity.actor == null ? <UserIcon /> : undefined}
      />
      <ListItemContent>
        <div className={header()}>
          <Text size={2}>{sentence}</Text>
          <Text size={1} color="faint">
            {dateTimeFormat(i18n.language, activity.createdAt)}
          </Text>
        </div>
      </ListItemContent>
    </ListItem>
  );
}
