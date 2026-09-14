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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { TaskLinearPublishField_task$key } from "#/__generated__/core/TaskLinearPublishField_task.graphql";
import type { TaskLinearPublishFieldMutation } from "#/__generated__/core/TaskLinearPublishFieldMutation.graphql";
import type { TaskLinearPublishFieldQuery } from "#/__generated__/core/TaskLinearPublishFieldQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { taskLinearPublishField } from "../variants";

const taskLinearPublishFieldFragment = graphql`
  fragment TaskLinearPublishField_task on Task {
    id
  }
`;

const taskLinearPublishFieldQuery = graphql`
  query TaskLinearPublishFieldQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        linearTeams {
          id
          name
          key
        }
      }
    }
  }
`;

const publishMutation = graphql`
  mutation TaskLinearPublishFieldMutation($input: PublishTaskToLinearInput!) {
    publishTaskToLinear(input: $input) {
      task {
        ...TaskLinearField_task
      }
    }
  }
`;

interface TaskLinearPublishFieldProps {
  taskKey: TaskLinearPublishField_task$key;
}

export function TaskLinearPublishField({ taskKey }: TaskLinearPublishFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const task = useFragment(taskLinearPublishFieldFragment, taskKey);
  const data = useLazyLoadQuery<TaskLinearPublishFieldQuery>(
    taskLinearPublishFieldQuery,
    { organizationId },
  );
  const [teamId, setTeamId] = useState<string | null>(null);
  const [publish, isPublishing] = useMutation<TaskLinearPublishFieldMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.linear.published"),
      errorToast: t("detailsPage.linear.errors.publish"),
    },
  );
  const { root } = taskLinearPublishField();

  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const teams = organization?.linearTeams ?? [];
  const selectedTeamId = teamId ?? teams[0]?.id ?? null;

  if (teams.length === 0) {
    return <Text size={2} color="faint">{t("detailsPage.empty")}</Text>;
  }

  return (
    <div className={root()}>
      <Select
        value={selectedTeamId}
        onValueChange={setTeamId}
        disabled={isPublishing}
      >
        <SelectTrigger size={1} aria-label={t("detailsPage.linear.team")}>
          {teams.find(team => team.id === selectedTeamId)?.name
            ?? t("detailsPage.linear.chooseTeam")}
        </SelectTrigger>
        <SelectPopup>
          {teams.map(team => (
            <SelectItem key={team.id} value={team.id}>
              {`${team.name} (${team.key})`}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
      <Button
        size={1}
        disabled={selectedTeamId == null || isPublishing}
        onClick={() => {
          if (selectedTeamId == null) {
            return;
          }

          void publish({
            variables: {
              input: {
                taskId: task.id,
                teamId: selectedTeamId,
              },
            },
          });
        }}
      >
        {t("detailsPage.linear.publish")}
      </Button>
    </div>
  );
}
