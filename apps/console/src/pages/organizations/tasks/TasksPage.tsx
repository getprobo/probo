// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, PageHeader } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { TasksCardOrganizationFragment$key } from "#/__generated__/core/TasksCardOrganizationFragment.graphql";
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
  const query = usePreloadedQuery(tasksPageQuery, queryRef);
  usePageTitle(__("Tasks"));

  return (
    <div className="space-y-6">
      <OrganizationTasksCard
        organizationRef={query.organization as TasksCardOrganizationFragment$key}
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
