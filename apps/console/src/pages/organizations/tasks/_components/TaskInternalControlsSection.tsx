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

import { XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { TaskInternalControlsSection_task$key } from "#/__generated__/core/TaskInternalControlsSection_task.graphql";
import type { TaskInternalControlsSectionQuery } from "#/__generated__/core/TaskInternalControlsSectionQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useUpdateTask } from "../_lib/useUpdateTask";
import { taskInternalControls } from "../variants";

import { InternalControlCombobox } from "./InternalControlCombobox";

const taskInternalControlsPageSize = 50;

const taskInternalControlsSectionFragment = graphql`
  fragment TaskInternalControlsSection_task on Task
  @refetchable(queryName: "TaskInternalControlsSectionQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    id
    canUpdate: permission(action: "core:task:update")
    internalControls(first: $first, after: $after)
      @connection(key: "useTaskInternalControlIds_internalControls") {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

interface TaskInternalControlsSectionProps {
  taskKey: TaskInternalControlsSection_task$key;
}

export function TaskInternalControlsSection({
  taskKey,
}: TaskInternalControlsSectionProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const {
    data: task,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    TaskInternalControlsSectionQuery,
    TaskInternalControlsSection_task$key
  >(taskInternalControlsSectionFragment, taskKey);
  const [updateTask, isUpdating] = useUpdateTask();
  const { root, names, name, control } = taskInternalControls();
  const internalControls = task.internalControls.edges.map(edge => edge.node);
  const linkedIds = new Set(internalControls.map(internalControl => internalControl.id));
  const linksIncomplete = hasNext || isLoadingNext;

  function save(internalControlIds: string[]) {
    void updateTask({
      previousInternalControlIds: internalControls.map(internalControl => internalControl.id),
      variables: {
        input: {
          taskId: task.id,
          internalControlIds,
        },
      },
    }, { successMessage: "" }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <Heading level={2} size={2} weight="medium">
          {t("detailsPage.linkInternalControls")}
        </Heading>
        <div className={names()}>
          {internalControls.length > 0
            ? internalControls.map(internalControl => (
                <div key={internalControl.id} className={name()}>
                  <Link
                    size={1}
                    title={internalControl.name}
                    className={control()}
                    to={`/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`}
                  >
                    {internalControl.name}
                  </Link>
                  {task.canUpdate && (
                    <IconButton
                      size={1}
                      variant="ghost"
                      color="neutral"
                      disabled={isUpdating || linksIncomplete}
                      className="shrink-0"
                      aria-label={t("detailsPage.actions.removeInternalControl")}
                      onClick={() => {
                        save(internalControls
                          .map(item => item.id)
                          .filter(id => id !== internalControl.id));
                      }}
                    >
                      <XIcon />
                    </IconButton>
                  )}
                </div>
              ))
            : (
                <Text size={1} color="faint">{t("detailsPage.none")}</Text>
              )}
          {hasNext && (
            <Button
              variant="ghost"
              color="neutral"
              size={1}
              className="self-start"
              loading={isLoadingNext}
              onClick={() => loadNext(taskInternalControlsPageSize)}
            >
              {t("detailsPage.actions.loadMoreInternalControls")}
            </Button>
          )}
          {task.canUpdate && (
            <InternalControlCombobox
              excludeIds={linkedIds}
              disabled={isUpdating || linksIncomplete}
              onSelect={(internalControl) => {
                save([...linkedIds, internalControl.id]);
              }}
            />
          )}
        </div>
      </div>
    </Card>
  );
}
