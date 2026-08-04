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
import { Breadcrumb, Button, Card, PageHeader } from "@probo/ui";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { NewMalaysiaPDPABreachPageCreateMutation } from "#/__generated__/core/NewMalaysiaPDPABreachPageCreateMutation.graphql";
import type { NewMalaysiaPDPABreachPageQuery } from "#/__generated__/core/NewMalaysiaPDPABreachPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { MalaysiaPDPABreachFormFields } from "./_components/MalaysiaPDPABreachFormFields";
import { MALAYSIA_PDPA_BREACH_CONNECTION_KEY } from "./_lib/breachDisplay";
import {
  type BreachFormErrors,
  type BreachFormField,
  type BreachFormValues,
  breachValuesToMutationFields,
  createBreachFormSchema,
  createDefaultBreachFormValues,
  toBreachFormErrors,
} from "./_lib/breachForm";

export const newMalaysiaPDPABreachPageQuery = graphql`
  query NewMalaysiaPDPABreachPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateMalaysiaPDPABreach: permission(
          action: "core:malaysia-pdpa-breach:create"
        )
      }
    }
  }
`;

const createMutation = graphql`
  mutation NewMalaysiaPDPABreachPageCreateMutation(
    $input: CreateMalaysiaPDPABreachIncidentInput!
    $connections: [ID!]!
  ) {
    createMalaysiaPDPABreachIncident(input: $input) {
      incidentEdge @prependEdge(connections: $connections) {
        node {
          id
          ...MalaysiaPDPABreachListItem_incident
        }
      }
    }
  }
`;

interface NewMalaysiaPDPABreachPageProps {
  queryRef: PreloadedQuery<NewMalaysiaPDPABreachPageQuery>;
}

export default function NewMalaysiaPDPABreachPage({
  queryRef,
}: NewMalaysiaPDPABreachPageProps) {
  const { t } = useTranslation("organizations/data-breaches");
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const data = usePreloadedQuery<NewMalaysiaPDPABreachPageQuery>(
    newMalaysiaPDPABreachPageQuery,
    queryRef,
  );
  const [values, setValues] = useState(createDefaultBreachFormValues);
  const [errors, setErrors] = useState<BreachFormErrors>({});
  const [createMalaysiaPDPABreach, isCreating]
    = useMutation<NewMalaysiaPDPABreachPageCreateMutation>(createMutation, {
      successMessage: t("messages.created"),
      errorToast: t("errors.create"),
    });

  usePageTitle(t("create.title"));

  if (data.organization?.__typename !== "Organization") {
    throw new Error("PAGE_NOT_FOUND: organization not found");
  }

  function updateValue<FieldName extends BreachFormField>(
    field: FieldName,
    value: BreachFormValues[FieldName],
  ) {
    setValues(current => ({ ...current, [field]: value }));
    setErrors(current => ({ ...current, [field]: undefined }));
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = createBreachFormSchema(t).safeParse(values);
    if (!result.success) {
      setErrors(toBreachFormErrors(result.error));
      return;
    }

    setErrors({});
    const input = result.data;
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      MALAYSIA_PDPA_BREACH_CONNECTION_KEY,
    );

    try {
      const response = await createMalaysiaPDPABreach({
        variables: {
          connections: [connectionId],
          input: {
            organizationId,
            ...breachValuesToMutationFields(input),
          },
        },
      });

      void navigate(
        `/organizations/${organizationId}/data-breaches/${response.createMalaysiaPDPABreachIncident.incidentEdge.node.id}`,
      );
    } catch {
      // useMutation already displays the localized server error.
    }
  }

  const listUrl = `/organizations/${organizationId}/data-breaches`;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          { label: t("list.title"), to: listUrl },
          { label: t("create.title") },
        ]}
      />
      <PageHeader
        title={t("create.title")}
        description={t("create.description")}
      />

      {!data.organization.canCreateMalaysiaPDPABreach
        ? (
            <Card padded>
              <p className="text-sm text-txt-secondary">
                {t("create.noPermission")}
              </p>
            </Card>
          )
        : (
            <form onSubmit={event => void onSubmit(event)} className="space-y-6">
              <MalaysiaPDPABreachFormFields
                values={values}
                errors={errors}
                disabled={isCreating}
                onValueChange={updateValue}
              />
              <div className="flex justify-end gap-3">
                <Button variant="secondary" to={listUrl} disabled={isCreating}>
                  {t("common.cancel")}
                </Button>
                <Button type="submit" disabled={isCreating}>
                  {isCreating
                    ? t("common.creating")
                    : t("common.create")}
                </Button>
              </div>
            </form>
          )}
    </div>
  );
}
