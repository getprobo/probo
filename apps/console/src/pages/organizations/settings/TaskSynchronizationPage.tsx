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

import { usePageTitle } from "@probo/hooks";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { TaskSynchronizationPageQuery } from "#/__generated__/core/TaskSynchronizationPageQuery.graphql";
import { TasksSettingsDialog } from "#/pages/organizations/tasks/_components/TasksSettingsDialog/TasksSettingsDialog";

export const taskSynchronizationPageQuery = graphql`
  query TaskSynchronizationPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...TasksSettingsDialog_organization
      }
    }
  }
`;

interface TaskSynchronizationPageProps {
  queryRef: PreloadedQuery<TaskSynchronizationPageQuery>;
}

export function TaskSynchronizationPage({ queryRef }: TaskSynchronizationPageProps) {
  const data = usePreloadedQuery<TaskSynchronizationPageQuery>(taskSynchronizationPageQuery, queryRef);
  const { t } = useTranslation();
  usePageTitle(t("nav.taskSynchronization"));

  if (data.organization?.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <div className="flex flex-col gap-6">
      <Heading level={1} size={6} weight="medium" highContrast>
        {t("nav.taskSynchronization")}
      </Heading>
      <TasksSettingsDialog organizationKey={data.organization} />
    </div>
  );
}
