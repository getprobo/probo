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

import type { FindingTasksSection_finding$key } from "#/__generated__/core/FindingTasksSection_finding.graphql";
import type { FindingTasksSectionLinkMutation } from "#/__generated__/core/FindingTasksSectionLinkMutation.graphql";
import type { FindingTasksSectionPaginationQuery } from "#/__generated__/core/FindingTasksSectionPaginationQuery.graphql";
import type { FindingTasksSectionUnlinkMutation } from "#/__generated__/core/FindingTasksSectionUnlinkMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { taskDetailsPath } from "#/pages/organizations/tasks/_lib/taskPath";

import { FindingTaskCombobox } from "./FindingTaskCombobox";
import { findingsPage } from "./variants";

const findingTasksSectionFragment = graphql`
  fragment FindingTasksSection_finding on Finding
  @refetchable(queryName: "FindingTasksSectionPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    id
    canLinkTask: permission(action: "core:finding:create-task-mapping")
    canUnlinkTask: permission(action: "core:finding:delete-task-mapping")
    tasks(first: $first, after: $after)
      @connection(key: "FindingTasksSection_tasks") {
      __id
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

const linkTaskMutation = graphql`
  mutation FindingTasksSectionLinkMutation(
    $input: CreateFindingTaskMappingInput!
    $connections: [ID!]!
  ) {
    createFindingTaskMapping(input: $input) {
      taskEdge @prependEdge(connections: $connections) {
        node {
          id
          name
        }
      }
    }
  }
`;

const unlinkTaskMutation = graphql`
  mutation FindingTasksSectionUnlinkMutation(
    $input: DeleteFindingTaskMappingInput!
    $connections: [ID!]!
  ) {
    deleteFindingTaskMapping(input: $input) {
      deletedTaskId @deleteEdge(connections: $connections)
    }
  }
`;

interface FindingTasksSectionProps {
  findingKey: FindingTasksSection_finding$key;
}

export function FindingTasksSection({ findingKey }: FindingTasksSectionProps) {
  const { t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const {
    data: finding,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    FindingTasksSectionPaginationQuery,
    FindingTasksSection_finding$key
  >(findingTasksSectionFragment, findingKey);
  const [linkTask, isLinking] = useMutation<FindingTasksSectionLinkMutation>(linkTaskMutation, {
    errorToast: t("details.tasks.linkError"),
  });
  const [unlinkTask, isUnlinking] = useMutation<FindingTasksSectionUnlinkMutation>(unlinkTaskMutation, {
    errorToast: t("details.tasks.unlinkError"),
  });
  const { links, linkRow, section } = findingsPage();
  const tasks = finding.tasks.edges.map(edge => edge.node);
  const linkedIds = new Set(tasks.map(task => task.id));
  const busy = isLinking || isUnlinking;

  return (
    <Card variant="soft" size={2}>
      <div className={section()}>
        <Heading level={2} size={2} weight="medium">
          {t("details.tasks.title")}
        </Heading>
        {finding.canLinkTask && (
          <FindingTaskCombobox
            excludeIds={linkedIds}
            disabled={busy}
            onSelect={(task) => {
              void linkTask({
                variables: {
                  input: { findingId: finding.id, taskId: task.id },
                  connections: [finding.tasks.__id],
                },
              }).catch(() => {
                // The mutation hook already reported the failure.
              });
            }}
          />
        )}
        {tasks.length === 0
          ? <Text size={2} color="faint">{t("details.tasks.empty")}</Text>
          : (
              <div className={links()}>
                {tasks.map(task => (
                  <div key={task.id} className={linkRow()}>
                    <Link to={taskDetailsPath(organizationId, task.id)} size={2}>
                      {task.name}
                    </Link>
                    {finding.canUnlinkTask && (
                      <IconButton
                        variant="ghost"
                        color="neutral"
                        aria-label={t("details.tasks.unlink")}
                        disabled={busy}
                        onClick={() => {
                          void unlinkTask({
                            variables: {
                              input: { findingId: finding.id, taskId: task.id },
                              connections: [finding.tasks.__id],
                            },
                          }).catch(() => {
                            // The mutation hook already reported the failure.
                          });
                        }}
                      >
                        <XIcon />
                      </IconButton>
                    )}
                  </div>
                ))}
              </div>
            )}
        {hasNext && (
          <Button
            variant="soft"
            color="neutral"
            disabled={isLoadingNext}
            onClick={() => loadNext(50)}
          >
            {t("details.tasks.loadMore")}
          </Button>
        )}
      </div>
    </Card>
  );
}
