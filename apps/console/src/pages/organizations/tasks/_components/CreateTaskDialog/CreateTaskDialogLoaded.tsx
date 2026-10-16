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

import { Field } from "@probo/ui/src/v2/form/Field";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import type { CreateTaskDialogLoadedQuery } from "#/__generated__/core/CreateTaskDialogLoadedQuery.graphql";

import { TaskLinearDraftField } from "../TaskLinearDraftField";

import { CreateTaskDialogFields } from "./CreateTaskDialogFields";
import type { CreateTaskDialogFormState } from "./useCreateTaskDialogForm";

export const linearConnectionQuery = graphql`
  query CreateTaskDialogLoadedQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        connectors(filter: { providers: [LINEAR_SYNC] }) {
          __typename
        }
        linearSettings {
          defaultTeam {
            id
          }
        }
        ...TaskLinearDraftField_organization
      }
    }
  }
`;

export interface CreateTaskDialogLoadedProps {
  queryRef: PreloadedQuery<CreateTaskDialogLoadedQuery>;
  form: CreateTaskDialogFormState;
}

export function CreateTaskDialogLoaded({
  queryRef,
  form,
}: CreateTaskDialogLoadedProps) {
  const { t } = useTranslation("organizations/tasks");
  const data = usePreloadedQuery<CreateTaskDialogLoadedQuery>(
    linearConnectionQuery,
    queryRef,
  );
  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const defaultTeamId = organization != null && organization.connectors.length > 0
    ? (organization.linearSettings.defaultTeam?.id ?? null)
    : null;

  return (
    <CreateTaskDialogFields
      form={form}
      linear={organization != null && organization.connectors.length > 0
        ? (
            <Field label={t("detailsPage.fields.linear")}>
              <TaskLinearDraftField
                organizationKey={organization}
                disabled={form.isSaving}
                onChange={form.setLinear}
                onIssue={form.applyLinearIssue}
              />
            </Field>
          )
        : null}
      onSubmit={() => {
        form.submit(form.linear ?? {
          teamId: defaultTeamId,
          issueId: null,
        });
      }}
    />
  );
}
