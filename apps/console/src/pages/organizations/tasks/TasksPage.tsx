// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, PageHeader } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import TaskFormDialog from "#/components/tasks/TaskFormDialog";
import { OrganizationTasksCard } from "#/components/tasks/TasksCard";

export const tasksPageQuery = graphql`
  query TasksPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...TasksCardOrganizationFragment
      }
    }
  }
`;

interface Props {
  queryRef: PreloadedQuery<TasksPageQuery>;
}

export default function TasksPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const query = usePreloadedQuery<TasksPageQuery>(tasksPageQuery, queryRef);
  usePageTitle(__("Tasks"));

  return (
    <div className="space-y-6">
      <OrganizationTasksCard
        organizationRef={query.organization}
        header={({ connectionId, canCreateTask, refetch }) => (
          <PageHeader
            title={__("Tasks")}
            description={__(
              "Track your assigned compliance tasks and keep progress on track.",
            )}
          >
            {canCreateTask && (
              <TaskFormDialog connection={connectionId} onCompleted={refetch}>
                <Button icon={IconPlusLarge}>{__("New task")}</Button>
              </TaskFormDialog>
            )}
          </PageHeader>
        )}
      />
    </div>
  );
}
