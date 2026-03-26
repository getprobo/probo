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
        header={({ connectionId, canCreateTask }) => (
          <PageHeader
            title={__("Tasks")}
            description={__(
              "Track your assigned compliance tasks and keep progress on track.",
            )}
          >
            {canCreateTask && (
              <TaskFormDialog connection={connectionId}>
                <Button icon={IconPlusLarge}>{__("New task")}</Button>
              </TaskFormDialog>
            )}
          </PageHeader>
        )}
      />
    </div>
  );
}
