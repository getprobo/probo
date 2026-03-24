import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Field,
  PageHeader,
} from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";
import { ConnectionHandler } from "relay-runtime";

import type { CreateAccessSourcePageQuery } from "#/__generated__/core/CreateAccessSourcePageQuery.graphql";
import type { CreateAccessSourceDialogMutation } from "#/__generated__/core/CreateAccessSourceDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { z } from "zod";

import { createAccessSourcePageQuery } from "./CreateAccessSourcePage";
import { createAccessSourceMutation } from "./dialogs/CreateAccessSourceDialog";

const csvSchema = z.object({
  name: z.string().min(1),
  csvData: z.string().min(1),
});

export default function CreateCsvAccessSourcePage({
  queryRef,
}: {
  queryRef: PreloadedQuery<CreateAccessSourcePageQuery>;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { register, handleSubmit }
    = useFormWithSchema(csvSchema, {
      defaultValues: {
        name: "",
        csvData: "",
      },
    });

  usePageTitle(__("Add CSV Access Source"));

  const { organization } = usePreloadedQuery(createAccessSourcePageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewSourcesTab_accessSources",
  );

  const [createAccessSource, isCreating]
    = useMutationWithToasts<CreateAccessSourceDialogMutation>(
      createAccessSourceMutation,
      {
        successMessage: __("Access source created successfully."),
        errorMessage: __("Failed to create access source"),
      },
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

  const onSubmit = async (data: z.infer<typeof csvSchema>) => {
    await createAccessSource({
      variables: {
        input: {
          organizationId,
          connectorId: null,
          name: data.name,
          csvData: data.csvData,
        },
        connections: connectionId ? [connectionId] : [],
      },
      onCompleted: () => {
        window.location.href = `/organizations/${organizationId}/access-reviews/sources`;
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
              <Link to={`/organizations/${organizationId}/access-reviews/sources/new`}>
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
