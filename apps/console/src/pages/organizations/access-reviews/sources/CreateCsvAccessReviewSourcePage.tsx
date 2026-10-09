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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { CreateCsvAccessReviewSourcePageQuery } from "#/__generated__/core/CreateCsvAccessReviewSourcePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { createAccessReviewSourcesMutation, prependCreatedSourceEdges } from "../dialogs/accessReviewSourceMutations";

import { csvSourcePage } from "./_components/variants";

export const createCsvAccessReviewSourcePageQuery = graphql`
  query CreateCsvAccessReviewSourcePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "access-review:source:create")
      }
    }
  }
`;

interface CreateCsvAccessReviewSourcePageProps {
  queryRef: PreloadedQuery<CreateCsvAccessReviewSourcePageQuery>;
}

export function CreateCsvAccessReviewSourcePage({
  queryRef,
}: CreateCsvAccessReviewSourcePageProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const toast = useToast();
  const organizationId = useOrganizationId();
  const { root, back, intro, form, actions } = csvSourcePage();
  const sourcesPath = `/organizations/${organizationId}/access-reviews/sources`;
  const newSourcePath = `${sourcesPath}/new`;

  usePageTitle(t("createCsvAccessReviewSourcePage.pageTitle"));

  const { organization } = usePreloadedQuery<CreateCsvAccessReviewSourcePageQuery>(
    createCsvAccessReviewSourcePageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewSourcesPage_accessReviewSources",
  );

  const [createAccessReviewSources, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourcesMutation,
      {
        errorToast: t("createCsvAccessReviewSourcePage.errors.create"),
      },
    );

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const name = data.get("name");
    const csvData = data.get("csvData");
    if (typeof name !== "string" || typeof csvData !== "string") {
      return;
    }

    void createAccessReviewSources({
      variables: {
        input: {
          organizationId,
          sources: [{
            connectorId: null,
            name,
            csvData,
          }],
        },
      },
      updater: (store) => {
        if (connectionId) {
          prependCreatedSourceEdges(store, connectionId);
        }
      },
    }).then(() => {
      toast.add({
        title: t("createCsvAccessReviewSourcePage.messages.created"),
        type: "success",
      });
      void navigate({ pathname: sourcesPath, search: location.search });
    }).catch(() => undefined);
  }

  return (
    <div className={root()}>
      <Link
        to={{ pathname: newSourcePath, search: location.search }}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("createCsvAccessReviewSourcePage.actions.back")}
      </Link>
      <div className={intro()}>
        <Heading level={1} size={6} weight="medium" highContrast>
          {t("createCsvAccessReviewSourcePage.title")}
        </Heading>
        <Text size={2} color="faint">
          {t("createCsvAccessReviewSourcePage.description")}
        </Text>
      </div>

      {organization.canCreateSource
        ? (
            <Card variant="soft" size={2}>
              <form className={form()} onSubmit={handleSubmit}>
                <Field
                  label={t("createCsvAccessReviewSourcePage.fields.name")}
                  required
                >
                  <TextField name="name" required />
                </Field>

                <Field
                  label={t("createCsvAccessReviewSourcePage.fields.csvData")}
                  required
                >
                  <Textarea
                    name="csvData"
                    required
                    rows={8}
                    placeholder={t("createCsvAccessReviewSourcePage.fields.csvPlaceholder")}
                  />
                </Field>
                <Text size={2} color="faint">
                  {t("createCsvAccessReviewSourcePage.supportedColumns")}
                </Text>

                <div className={actions()}>
                  <Button type="submit" loading={isCreating}>
                    {t("createCsvAccessReviewSourcePage.actions.create")}
                  </Button>
                </div>
              </form>
            </Card>
          )
        : (
            <Card variant="soft" size={2}>
              <Text size={2} color="faint">
                {t("createCsvAccessReviewSourcePage.permissionDenied")}
              </Text>
            </Card>
          )}
    </div>
  );
}
