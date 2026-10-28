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
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { TasksPageList } from "./_components/TasksPageList/TasksPageList";
import { TasksSettingsDialog } from "./_components/TasksSettingsDialog/TasksSettingsDialog";
import { taskNewPath } from "./_lib/taskPath";
import { tasksPage } from "./variants";

export const tasksPageQuery = graphql`
  query TasksPageQuery(
    $organizationId: ID!
    $query: String
    $assignedToId: ID
    $internalControlId: ID
    $deadlineOnOrBefore: Datetime
    $deadlineOnOrAfter: Datetime
    $hasDeadline: Boolean
    $hasAssignee: Boolean
    $includeInProgress: Boolean!
    $includeTodo: Boolean!
    $includeBacklog: Boolean!
    $includeDone: Boolean!
    $includeCanceled: Boolean!
    $includeDuplicate: Boolean!
  ) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateTask: permission(action: "core:task:create")
        ...TasksPageList_organization
          @arguments(
            query: $query
            assignedToId: $assignedToId
            internalControlId: $internalControlId
            deadlineOnOrBefore: $deadlineOnOrBefore
            deadlineOnOrAfter: $deadlineOnOrAfter
            hasDeadline: $hasDeadline
            hasAssignee: $hasAssignee
            includeInProgress: $includeInProgress
            includeTodo: $includeTodo
            includeBacklog: $includeBacklog
            includeDone: $includeDone
            includeCanceled: $includeCanceled
            includeDuplicate: $includeDuplicate
          )
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
  const organizationId = useOrganizationId();
  const location = useLocation();
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
          {organization.canCreateTask && (
            <ButtonLink
              to={{ pathname: taskNewPath(organizationId), search: location.search }}
              variant="solid"
              iconStart={<PlusIcon />}
            >
              {t("tasks.actions.create")}
            </ButtonLink>
          )}
          <TasksSettingsDialog organizationKey={organization} />
        </div>
      </div>
      <TasksPageList organizationKey={organization} />
    </div>
  );
}
