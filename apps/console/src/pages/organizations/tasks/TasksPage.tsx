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

import { PlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { PageHeader } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import { OrganizationTasksCard } from "#/components/tasks/TasksCard";

import { CreateTaskDialog } from "./_components/CreateTaskDialog";
import { TasksSettingsDialog } from "./_components/TasksSettingsDialog";
import { tasksPage } from "./variants";

export const tasksPageQuery = graphql`
  query TasksPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...TasksCardOrganizationFragment
        ...TasksSettingsDialog_organization
      }
    }
  }
`;

interface TasksPageProps {
  queryRef: PreloadedQuery<TasksPageQuery>;
}

export function TasksPage({ queryRef }: TasksPageProps) {
  const { t } = useTranslation();
  const query = usePreloadedQuery<TasksPageQuery>(tasksPageQuery, queryRef);
  usePageTitle(t("tasks.title"));

  const organization = query.organization.__typename === "Organization"
    ? query.organization
    : null;
  const { root, actions } = tasksPage();

  if (organization == null) {
    return (
      <div className={root()}>
        <PageHeader
          title={t("tasks.title")}
          description={t("tasks.description")}
        />
      </div>
    );
  }

  return (
    <div className={root()}>
      <OrganizationTasksCard
        organizationRef={organization}
        header={({ connectionId, canCreateTask, refetch }) => (
          <PageHeader
            title={t("tasks.title")}
            description={t("tasks.description")}
          >
            <div className={actions()}>
              <TasksSettingsDialog organizationKey={organization} />
              {canCreateTask && (
                <CreateTaskDialog connectionId={connectionId} onCompleted={refetch}>
                  <Button
                    variant="solid"
                    color="neutral"
                    highContrast
                    iconStart={<PlusIcon />}
                  >
                    {t("tasks.actions.create")}
                  </Button>
                </CreateTaskDialog>
              )}
            </div>
          </PageHeader>
        )}
      />
    </div>
  );
}
