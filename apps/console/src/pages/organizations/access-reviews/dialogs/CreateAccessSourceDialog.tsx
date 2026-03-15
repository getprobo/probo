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
} from "@probo/ui";
import { type ReactNode, useEffect, useMemo } from "react";
import { Controller, useWatch } from "react-hook-form";
import { graphql } from "react-relay";
import { useSearchParams } from "react-router";
import { z } from "zod";

import type { CreateAccessSourceDialogMutation } from "#/__generated__/core/CreateAccessSourceDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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
  accessReviewId: string;
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
  provider: z.enum(["GOOGLE_WORKSPACE", "LINEAR"]).optional(),
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

function providerLabel(provider: "GOOGLE_WORKSPACE" | "LINEAR") {
  switch (provider) {
    case "GOOGLE_WORKSPACE":
      return "Google Workspace";
    case "LINEAR":
      return "Linear";
    default:
      return provider;
  }
}

export function CreateAccessSourceDialog({
  children,
  organizationId,
  accessReviewId,
  connectionId,
  connectors,
  preselectedConnectorId,
}: Props) {
  const { __ } = useTranslate();
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
          provider:
            preselectedConnector?.provider === "GOOGLE_WORKSPACE"
            || preselectedConnector?.provider === "LINEAR"
              ? preselectedConnector.provider
              : "GOOGLE_WORKSPACE",
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
    () =>
      connectors.filter(
        connector =>
          connector.provider === "GOOGLE_WORKSPACE"
          || connector.provider === "LINEAR",
      ),
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
    if (preselectedConnector.provider !== "GOOGLE_WORKSPACE" && preselectedConnector.provider !== "LINEAR") return;
    setValue("sourceType", "OAUTH2");
    setValue("provider", preselectedConnector.provider);
    setValue("connectorId", preselectedConnector.id);
  }, [preselectedConnector, setValue]);

  const [createAccessSource, isCreating]
    = useMutationWithToasts<CreateAccessSourceDialogMutation>(
      createAccessSourceMutation,
      {
        successMessage: __("Access source created successfully."),
        errorMessage: __("Failed to create access source"),
      },
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

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const isOAuth = data.sourceType === "OAUTH2";
    await createAccessSource({
      variables: {
        input: {
          accessReviewId,
          connectorId: isOAuth ? data.connectorId : null,
          name: data.name,
          csvData: isOAuth ? null : (data.csvData || null),
        },
        connections: [connectionId],
      },
      onCompleted: () => {
        clearConnectorQueryParam();
        reset();
        ref.current?.close();
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
                          {providerLabel(connector.provider === "GOOGLE_WORKSPACE" ? "GOOGLE_WORKSPACE" : "LINEAR")}
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
