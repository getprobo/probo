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
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, useEffect, useMemo } from "react";
import { Controller, useWatch } from "react-hook-form";
import { graphql, useMutation } from "react-relay";
import { useSearchParams } from "react-router";
import { z } from "zod";

import type { CreateAccessSourceDialogMutation } from "#/__generated__/core/CreateAccessSourceDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

export const createAccessSourceMutation = graphql`
  mutation CreateAccessSourceDialogMutation(
    $input: CreateAccessSourceInput!
    $connections: [ID!]!
  ) {
    createAccessSource(input: $input) {
      accessSourceEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          createdAt
          ...AccessSourceRowFragment
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  organizationId: string;
  connectionId: string;
  connectors: ReadonlyArray<{
    readonly id: string;
    readonly provider: "GOOGLE_WORKSPACE" | "LINEAR" | "SLACK";
    readonly createdAt: string;
  }>;
  preselectedConnectorId: string | null;
};

const schema = z.object({
  name: z.string().min(1),
  sourceType: z.enum(["CSV", "OAUTH2"]),
  provider: z.enum(["GOOGLE_WORKSPACE", "LINEAR", "SLACK"]).optional(),
  connectorId: z.string().optional(),
  csvData: z.string().optional(),
}).superRefine((data, ctx) => {
  if (data.sourceType === "CSV" && !data.csvData?.trim()) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["csvData"],
      message: "CSV data is required for CSV sources.",
    });
  }

  if (data.sourceType === "OAUTH2") {
    if (!data.provider) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["provider"],
        message: "Provider is required for OAuth2 sources.",
      });
    }
    if (!data.connectorId) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["connectorId"],
        message: "Connector is required for OAuth2 sources.",
      });
    }
  }
});

function providerLabel(provider: "GOOGLE_WORKSPACE" | "LINEAR" | "SLACK") {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return "Google Workspace";
    case "LINEAR":
      return "Linear";
    case "SLACK":
      return "Slack";
    default:
      return provider;
  }
}

export function CreateAccessSourceDialog({
  children,
  organizationId,
  connectionId,
  connectors,
  preselectedConnectorId,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const preselectedConnector = useMemo(
    () => connectors.find(connector => connector.id === preselectedConnectorId),
    [connectors, preselectedConnectorId],
  );
  const { control, register, handleSubmit, reset, setValue }
    = useFormWithSchema(
      schema,
      {
        defaultValues: {
          name: "",
          sourceType: preselectedConnector ? "OAUTH2" : "CSV",
          provider: preselectedConnector?.provider ?? "GOOGLE_WORKSPACE",
          connectorId: preselectedConnector?.id,
          csvData: "",
        },
      },
    );
  const sourceType = useWatch({ control, name: "sourceType" });
  const provider = useWatch({ control, name: "provider" });
  const connectorId = useWatch({ control, name: "connectorId" });
  const ref = useDialogRef();

  const providerConnectors = useMemo(
    () => connectors,
    [connectors],
  );
  const selectableConnectors = useMemo(
    () =>
      providerConnectors.filter(
        connector => !provider || connector.provider === provider,
      ),
    [provider, providerConnectors],
  );

  useEffect(() => {
    if (!provider) {
      setValue("connectorId", undefined);
      return;
    }
    if (
      connectorId
      && !selectableConnectors.some(connector => connector.id === connectorId)
    ) {
      setValue("connectorId", undefined);
    }
  }, [provider, connectorId, selectableConnectors, setValue]);

  useEffect(() => {
    if (!preselectedConnector) return;
    setValue("sourceType", "OAUTH2");
    setValue("provider", preselectedConnector.provider);
    setValue("connectorId", preselectedConnector.id);
  }, [preselectedConnector, setValue]);

  const [createAccessSource, isCreating]
    = useMutation<CreateAccessSourceDialogMutation>(
      createAccessSourceMutation,
    );

  const clearConnectorQueryParam = () => {
    if (!searchParams.get("connector_id")) {
      return;
    }
    setSearchParams((params) => {
      params.delete("connector_id");
      return params;
    });
  };

  const startOAuthConnection = () => {
    if (!provider) {
      return;
    }

    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseURL);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", provider);
    url.searchParams.append("continue", `/organizations/${organizationId}/access-reviews`);
    window.location.href = url.toString();
  };

  const onSubmit = (data: z.infer<typeof schema>) => {
    const isOAuth = data.sourceType === "OAUTH2";
    createAccessSource({
      variables: {
        input: {
          organizationId,
          connectorId: isOAuth ? data.connectorId : null,
          name: data.name,
          csvData: isOAuth ? null : (data.csvData || null),
        },
        connections: [connectionId],
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
        clearConnectorQueryParam();
        reset();
        ref.current?.close();
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
    <Dialog
      ref={ref}
      trigger={children}
      title={(
        <Breadcrumb
          items={[
            __("Access Reviews"),
            __("New Access Source"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />
          <Field label={__("Source type")}>
            <Controller
              control={control}
              name="sourceType"
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <Option value="CSV">{__("CSV Upload")}</Option>
                  <Option value="OAUTH2">{__("OAuth2 Connector")}</Option>
                </Select>
              )}
            />
          </Field>
          {sourceType === "OAUTH2" && (
            <>
              <Field label={__("Provider")}>
                <Controller
                  control={control}
                  name="provider"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={(value) => {
                        field.onChange(value);
                        setValue("connectorId", undefined);
                      }}
                    >
                      <Option value="GOOGLE_WORKSPACE">{__("Google Workspace")}</Option>
                      <Option value="LINEAR">{__("Linear")}</Option>
                      <Option value="SLACK">{__("Slack")}</Option>
                    </Select>
                  )}
                />
              </Field>
              <Field label={__("Connector")}>
                <Controller
                  control={control}
                  name="connectorId"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {selectableConnectors.map(connector => (
                        <Option key={connector.id} value={connector.id}>
                          {`${providerLabel(connector.provider)} (${new Date(connector.createdAt).toLocaleDateString(undefined, { month: "short", year: "numeric" })})`}
                        </Option>
                      ))}
                    </Select>
                  )}
                />
              </Field>
              <div className="flex items-center justify-between gap-3">
                <p className="text-txt-secondary text-sm">
                  {__("Need a new connection? Connect your provider and come back to continue creating this source.")}
                </p>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={startOAuthConnection}
                  disabled={!provider}
                >
                  {__("Connect")}
                </Button>
              </div>
            </>
          )}
          {sourceType === "CSV" && (
            <>
              <Field
                label={__("CSV Data")}
                {...register("csvData")}
                type="textarea"
                placeholder="email,full_name,role,job_title,is_admin,active,external_id"
              />
              <p className="text-txt-secondary text-sm">
                {__("Paste CSV content with a header row. Supported columns: email, full_name, role, job_title, is_admin, active, external_id.")}
              </p>
            </>
          )}
        </DialogContent>
        <DialogFooter>
          <Button disabled={isCreating} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
