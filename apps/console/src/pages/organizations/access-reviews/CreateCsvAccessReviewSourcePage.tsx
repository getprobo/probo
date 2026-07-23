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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  Field,
  PageHeader,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";
import { z } from "zod";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { CreateCsvAccessReviewSourcePageQuery } from "#/__generated__/core/CreateCsvAccessReviewSourcePageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { createAccessReviewSourceMutation } from "./dialogs/accessReviewSourceMutations";

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

const csvSchema = z.object({
  name: z.string().min(1),
  csvData: z.string().min(1),
});

export default function CreateCsvAccessReviewSourcePage({
  queryRef,
}: {
  queryRef: PreloadedQuery<CreateCsvAccessReviewSourcePageQuery>;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const { register, handleSubmit }
    = useFormWithSchema(csvSchema, {
      defaultValues: {
        name: "",
        csvData: "",
      },
    });

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
    "AccessReviewSourcesTab_accessReviewSources",
  );

  const [createAccessReviewSource, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourceMutation,
    );

  if (!organization.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {t("createCsvAccessReviewSourcePage.permissionDenied")}
        </p>
      </Card>
    );
  }

  const onSubmit = (data: z.infer<typeof csvSchema>) => {
    createAccessReviewSource({
      variables: {
        input: {
          organizationId,
          connectorId: null,
          name: data.name,
          csvData: data.csvData,
        },
        connections: connectionId ? [connectionId] : [],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("createCsvAccessReviewSourcePage.messages.error"),
            description: formatError(
              t("createCsvAccessReviewSourcePage.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("createCsvAccessReviewSourcePage.messages.success"),
          description: t("createCsvAccessReviewSourcePage.messages.created"),
          variant: "success",
        });
        void navigate(`/organizations/${organizationId}/access-reviews/sources`);
      },
      onError(error) {
        toast({
          title: t("createCsvAccessReviewSourcePage.messages.error"),
          description: formatError(
            t("createCsvAccessReviewSourcePage.errors.create"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("createCsvAccessReviewSourcePage.title")}
        description={t("createCsvAccessReviewSourcePage.description")}
      />

      <Card padded>
        <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
          <Field
            label={t("createCsvAccessReviewSourcePage.fields.name")}
            {...register("name")}
            type="text"
            required
          />

          <Field
            label={t("createCsvAccessReviewSourcePage.fields.csvData")}
            {...register("csvData")}
            type="textarea"
            placeholder={t("createCsvAccessReviewSourcePage.fields.csvPlaceholder")}
            required
          />
          <p className="text-txt-secondary text-sm">
            {t("createCsvAccessReviewSourcePage.supportedColumns")}
          </p>

          <div className="flex items-center justify-end gap-2">
            <Button variant="secondary" asChild>
              <Link to={`/organizations/${organizationId}/access-reviews/sources`}>
                {t("createCsvAccessReviewSourcePage.actions.back")}
              </Link>
            </Button>
            <Button disabled={isCreating} type="submit">
              {t("createCsvAccessReviewSourcePage.actions.create")}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
