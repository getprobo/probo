// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Field,
  PageHeader,
  useToast,
} from "@probo/ui";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateAccessSourceDialogMutation } from "#/__generated__/core/CreateAccessSourceDialogMutation.graphql";
import type { CreateCsvAccessSourcePageQuery } from "#/__generated__/core/CreateCsvAccessSourcePageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { createAccessSourceMutation } from "./dialogs/CreateAccessSourceDialog";

export const createCsvAccessSourcePageQuery = graphql`
  query CreateCsvAccessSourcePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "core:access-source:create")
      }
    }
  }
`;

const csvSchema = z.object({
  name: z.string().min(1),
  csvData: z.string().min(1),
});

export default function CreateCsvAccessSourcePage({
  queryRef,
}: {
  queryRef: PreloadedQuery<CreateCsvAccessSourcePageQuery>;
}) {
  const { __ } = useTranslate();
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

  usePageTitle(__("Add CSV Access Source"));

  const { organization } = usePreloadedQuery(createCsvAccessSourcePageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewSourcesTab_accessSources",
  );

  const [createAccessSource, isCreating]
    = useMutation<CreateAccessSourceDialogMutation>(
      createAccessSourceMutation,
    );

  if (!organization.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {__("You do not have permission to create access sources.")}
        </p>
      </Card>
    );
  }

  const onSubmit = (data: z.infer<typeof csvSchema>) => {
    createAccessSource({
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
            title: __("Error"),
            description: formatError(
              __("Failed to create access source"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access source created successfully."),
          variant: "success",
        });
        void navigate(`/organizations/${organizationId}/access-reviews/sources`);
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create access source"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Add CSV access source")}
        description={__(
          "Paste CSV content with a header row. This source will be saved and available in Access Reviews.",
        )}
      />

      <Card padded>
        <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />

          <Field
            label={__("CSV Data")}
            {...register("csvData")}
            type="textarea"
            placeholder="email,full_name,role,job_title,is_admin,active,external_id"
            required
          />
          <p className="text-txt-secondary text-sm">
            {__("Supported columns: email, full_name, role, job_title, is_admin, active, external_id.")}
          </p>

          <div className="flex items-center justify-end gap-2">
            <Button variant="secondary" asChild>
              <Link to={`/organizations/${organizationId}/access-reviews/sources`}>
                {__("Back")}
              </Link>
            </Button>
            <Button disabled={isCreating} type="submit">
              {__("Create")}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
