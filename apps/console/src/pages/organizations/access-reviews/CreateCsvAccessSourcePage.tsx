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
  Dropzone,
  Field,
  IconCrossLargeX,
  PageHeader,
  useToast,
} from "@probo/ui";
import { useCallback, useState } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";
import { z } from "zod";

import type { accessSourceMutationsCreateMutation } from "#/__generated__/core/accessSourceMutationsCreateMutation.graphql";
import type { CreateCsvAccessSourcePageQuery } from "#/__generated__/core/CreateCsvAccessSourcePageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { createAccessSourceMutation } from "./dialogs/accessSourceMutations";

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
});

const csvAccept = {
  "text/csv": [".csv"],
  "application/csv": [".csv"],
};

const MAX_CSV_MB = 10;

export default function CreateCsvAccessSourcePage({
  queryRef,
}: {
  queryRef: PreloadedQuery<CreateCsvAccessSourcePageQuery>;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [csvContent, setCsvContent] = useState<string | null>(null);
  const [csvError, setCsvError] = useState<string | null>(null);
  const [isReadingFile, setIsReadingFile] = useState(false);

  const { register, handleSubmit, setValue, getValues }
    = useFormWithSchema(csvSchema, {
      defaultValues: {
        name: "",
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
    = useMutation<accessSourceMutationsCreateMutation>(
      createAccessSourceMutation,
    );

  const handleFileDrop = useCallback(
    (acceptedFiles: File[]) => {
      const file = acceptedFiles[0];
      if (!file) {
        return;
      }

      setCsvError(null);
      setCsvFile(file);
      setCsvContent(null);
      setIsReadingFile(true);

      file.text()
        .then((text) => {
          if (text.trim().length === 0) {
            setCsvError(__("The selected file is empty."));
            setCsvContent(null);
            return;
          }
          setCsvContent(text);
          // Suggest a name from the file name (without extension), only if the
          // user has not already typed something.
          if (!getValues("name")) {
            setValue("name", file.name.replace(/\.[^/.]+$/, ""), {
              shouldValidate: true,
            });
          }
        })
        .catch(() => {
          setCsvError(__("Could not read the selected file."));
          setCsvContent(null);
        })
        .finally(() => {
          setIsReadingFile(false);
        });
    },
    [__, getValues, setValue],
  );

  const handleClearFile = () => {
    setCsvFile(null);
    setCsvContent(null);
    setCsvError(null);
  };

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
    if (!csvContent) {
      setCsvError(__("Please select a CSV file."));
      return;
    }

    createAccessSource({
      variables: {
        input: {
          organizationId,
          connectorId: null,
          name: data.name,
          csvData: csvContent,
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

  const canSubmit = !!csvContent && !csvError && !isReadingFile && !isCreating;

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Add CSV access source")}
        description={__(
          "Upload a CSV file with a header row. This source will be saved and available in Access Reviews.",
        )}
      />

      <Card padded>
        <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
          <div className="space-y-2">
            <Dropzone
              description={__(
                "Drop a .csv file here or click to browse (max 10MB).",
              )}
              isUploading={isReadingFile}
              onDrop={handleFileDrop}
              accept={csvAccept}
              maxSize={MAX_CSV_MB}
            />
            {csvFile && !csvError && (
              <div className="flex items-center justify-between rounded-md border border-border-low bg-subtle px-3 py-2 text-sm">
                <div className="flex flex-col">
                  <span className="font-medium">{csvFile.name}</span>
                  {csvContent && (
                    <span className="text-txt-tertiary text-xs">
                      {formatFileSummary(__, csvFile.size, csvContent)}
                    </span>
                  )}
                </div>
                <Button
                  type="button"
                  variant="quaternary"
                  icon={IconCrossLargeX}
                  onClick={handleClearFile}
                  disabled={isCreating || isReadingFile}
                >
                  {__("Remove")}
                </Button>
              </div>
            )}
            {csvError && (
              <p className="text-sm text-txt-danger">{csvError}</p>
            )}
            <p className="text-txt-secondary text-sm">
              {__(
                "Supported columns: email, full_name, role, job_title, is_admin, active, external_id.",
              )}
            </p>
          </div>

          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />

          <div className="flex items-center justify-end gap-2">
            <Button variant="secondary" asChild>
              <Link to={`/organizations/${organizationId}/access-reviews/sources`}>
                {__("Back")}
              </Link>
            </Button>
            <Button disabled={!canSubmit} type="submit">
              {__("Create")}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}

function formatFileSummary(
  __: (s: string) => string,
  bytes: number,
  csv: string,
): string {
  const sizeKb = Math.max(1, Math.round(bytes / 1024));
  const lines = csv.split(/\r?\n/).filter(line => line.length > 0).length;
  const dataRows = Math.max(0, lines - 1);
  return `${sizeKb} KB · ${dataRows} ${__("rows")}`;
}
