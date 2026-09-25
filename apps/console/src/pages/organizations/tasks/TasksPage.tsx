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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import { OrganizationTasksCard } from "#/components/tasks/TasksCard";

import { CreateTaskDialog } from "./_components/CreateTaskDialog";
import { TasksSettingsDialog } from "./_components/TasksSettingsDialog";
import { tasksPage } from "./variants";

export const tasksPageQuery = graphql`
  query TasksPageQuery($organizationId: ID!, $filter: TaskFilter) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...TasksCardOrganizationFragment @arguments(filter: $filter)
        ...TasksSettingsDialog_organization
      }
    }
  }
`;

interface TasksPageProps {
  queryRef: PreloadedQuery<TasksPageQuery>;
  onReady?: () => void;
}

export function TasksPage({ queryRef, onReady }: TasksPageProps) {
  const { t } = useTranslation();
  const query = usePreloadedQuery<TasksPageQuery>(tasksPageQuery, queryRef);
  usePageTitle(t("tasks.title"));
  useEffect(() => {
    onReady?.();
  }, [onReady]);

  const { organization } = query;
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const { root, header, intro, actions } = tasksPage();

  return (
    <div className={root()}>
      <OrganizationTasksCard
        organizationRef={organization}
        header={({ connectionId, canCreateTask, refetch }) => (
          <div className={header()}>
            <div className={intro()}>
              <Heading level={1} size={6} weight="medium" highContrast>
                {t("tasks.title")}
              </Heading>
              <Text size={2} color="faint">
                {t("tasks.description")}
              </Text>
            </div>
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
          </div>
        )}
      />
    </div>
  );
}
