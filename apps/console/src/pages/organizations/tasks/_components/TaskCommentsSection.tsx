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

import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { Suspense } from "react";
import { graphql, useFragment } from "react-relay";

import type { TaskCommentsSection_task$key } from "#/__generated__/core/TaskCommentsSection_task.graphql";

import { taskCommentsSection } from "../variants";

import { TaskCommentsList } from "./TaskCommentsList";

const taskCommentsSectionFragment = graphql`
  fragment TaskCommentsSection_task on Task {
    canListComments: permission(action: "core:task-comment:list")
  }
`;

interface TaskCommentsSectionProps {
  taskKey: TaskCommentsSection_task$key;
}

function TaskCommentsSectionFallback() {
  const { root, header } = taskCommentsSection();

  return (
    <div className={root()}>
      <div className={header()}>
        <HeadingSkeleton size={4} className="w-32" />
      </div>
      <ListSkeleton count={2} />
    </div>
  );
}

export function TaskCommentsSection({ taskKey }: TaskCommentsSectionProps) {
  const task = useFragment(taskCommentsSectionFragment, taskKey);

  if (!task.canListComments) {
    return null;
  }

  return (
    <Suspense fallback={<TaskCommentsSectionFallback />}>
      <TaskCommentsList />
    </Suspense>
  );
}
