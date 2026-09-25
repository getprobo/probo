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

import { CopyIcon } from "@phosphor-icons/react";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import type {
  AzureEnvironment,
  WorkloadIdentityFormCreateMutation,
} from "#/__generated__/core/WorkloadIdentityFormCreateMutation.graphql";
import type { WorkloadIdentityFormDeleteMutation } from "#/__generated__/core/WorkloadIdentityFormDeleteMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import {
  isAWSRoleARN,
  isAzureGUID,
  isGCPServiceAccountEmail,
  isGCPWorkloadIdentityProvider,
} from "#/pages/organizations/access-reviews/dialogs/_lib/connectorSettings";

import { integrationListPath } from "../_lib/integrationPath";
import { ConnectFormFooter } from "./ConnectFormFooter";

const azureEnvironments = [
  "AZURE_PUBLIC",
  "AZURE_GOVERNMENT",
  "AZURE_GOVERNMENT_DOD",
  "AZURE_CHINA",
] as const satisfies ReadonlyArray<AzureEnvironment>;

const createWorkloadIdentityConnectorMutation = graphql`
  mutation WorkloadIdentityFormCreateMutation($input: CreateWorkloadIdentityConnectorInput!) {
    createWorkloadIdentityConnector(input: $input) {
      connector {
        id
        connectionStatus
      }
    }
  }
`;

const deleteConnectorMutation = graphql`
  mutation WorkloadIdentityFormDeleteMutation($input: DeleteConnectorInput!) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

type DriverProvider = ConnectVendorPageQuery["response"]["accessReviewDrivers"][number]["provider"];

interface WorkloadIdentityFormProps {
  organizationId: string;
  provider: DriverProvider;
  documentationUrl: string | null | undefined;
  awsConnectorSetup: ConnectVendorPageQuery["response"]["awsConnectorSetup"];
  gcpConnectorSetup: ConnectVendorPageQuery["response"]["gcpConnectorSetup"];
  azureConnectorSetup: ConnectVendorPageQuery["response"]["azureConnectorSetup"];
}

export function WorkloadIdentityForm({
  organizationId,
  provider,
  documentationUrl,
  awsConnectorSetup,
  gcpConnectorSetup,
  azureConnectorSetup,
}: WorkloadIdentityFormProps) {
  if (provider === "AWS") {
    return (
      <AwsWorkloadIdentityForm
        organizationId={organizationId}
        documentationUrl={documentationUrl}
        setup={awsConnectorSetup}
      />
    );
  }
  if (provider === "GCP") {
    return (
      <GcpWorkloadIdentityForm
        organizationId={organizationId}
        documentationUrl={documentationUrl}
        setup={gcpConnectorSetup}
      />
    );
  }
  if (provider === "AZURE") {
    return (
      <AzureWorkloadIdentityForm
        organizationId={organizationId}
        documentationUrl={documentationUrl}
        setup={azureConnectorSetup}
      />
    );
  }

  return null;
}

function useCopyValue() {
  const { toast } = useToast();

  return (value: string, title: string, failure: string) => {
    const onCopyFailure = () => {
      toast({ title: failure, description: failure, variant: "error" });
    };
    if (!navigator.clipboard?.writeText) {
      onCopyFailure();
      return;
    }
    navigator.clipboard.writeText(value).then(
      () => toast({ title, description: title, variant: "success" }),
      onCopyFailure,
    );
  };
}

const setupFields = ["issuer", "audience", "subject"] as const;

function setupRows(
  t: (key: string) => string,
  pageKey: string,
  setup: { issuer: string; audience: string; subject: string },
  copyValue: (value: string, title: string, failure: string) => void,
) {
  const copiedKey = {
    issuer: "copiedIssuer",
    audience: "copiedAudience",
    subject: "copiedSubject",
  } as const;

  return setupFields.map(field => ({
    label: t(`${pageKey}.fields.${field}`),
    value: setup[field],
    copyLabel: t(`${pageKey}.actions.copy`),
    onCopy: () => copyValue(
      setup[field],
      t(`${pageKey}.messages.${copiedKey[field]}`),
      t(`${pageKey}.messages.copyFailed`),
    ),
  }));
}

function useFinishWorkloadIdentity(organizationId: string) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const navigate = useNavigate();
  const [isCreating, setIsCreating] = useState(false);
  const [createConnector] = useMutation<WorkloadIdentityFormCreateMutation>(createWorkloadIdentityConnectorMutation);
  const [deleteConnector] = useMutation<WorkloadIdentityFormDeleteMutation>(deleteConnectorMutation);

  async function finish(
    input: WorkloadIdentityFormCreateMutation["variables"]["input"],
    errors: { create: string; disconnected: string; delete: string; errorTitle: string },
  ) {
    if (isCreating) {
      return;
    }
    setIsCreating(true);
    try {
      const created = await createConnector({
        variables: { input },
      }, { errorToast: errors.create });
      const connector = created.createWorkloadIdentityConnector.connector;
      if (connector.connectionStatus !== "CONNECTED") {
        toast({
          title: errors.errorTitle,
          description: errors.disconnected,
          variant: "error",
        });
        await deleteConnector({
          variables: { input: { connectorId: connector.id } },
        }, { errorToast: errors.delete });
        return;
      }
      toast({
        title: t("marketplacePage.connected"),
        description: t("listPage.messages.connectedDescription"),
        variant: "success",
      });
      void navigate(integrationListPath(organizationId));
    } catch {
      return;
    } finally {
      setIsCreating(false);
    }
  }

  return { finish, isCreating };
}

function SetupValues({
  rows,
}: {
  rows: Array<{ label: string; value: string; copyLabel: string; onCopy: () => void }>;
}) {
  return (
    <div className="flex flex-col gap-2 rounded-2 bg-sand-3 p-3">
      {rows.map(row => (
        <div key={row.label} className="flex flex-col gap-1">
          <Text size={1} color="faint">{row.label}</Text>
          <div className="flex min-w-0 items-center gap-1">
            <Code size={2} className="min-w-0 flex-1 break-all">{row.value}</Code>
            <IconButton
              type="button"
              size={1}
              variant="soft"
              color="neutral"
              aria-label={row.copyLabel}
              onClick={row.onCopy}
            >
              <CopyIcon />
            </IconButton>
          </div>
        </div>
      ))}
    </div>
  );
}

export function WorkloadIdentityInstallActions({
  provider,
  awsConnectorSetup,
  gcpConnectorSetup,
  azureConnectorSetup,
}: Pick<WorkloadIdentityFormProps, "provider" | "awsConnectorSetup" | "gcpConnectorSetup" | "azureConnectorSetup">) {
  const { t } = useTranslation("organizations/access-reviews");
  const copyValue = useCopyValue();

  if (provider === "AWS") {
    return (
      <div className="flex shrink-0 flex-wrap justify-end gap-2">
        {awsConnectorSetup.cloudFormationQuickCreateURL != null && (
          <ButtonAnchor href={awsConnectorSetup.cloudFormationQuickCreateURL} target="_blank" rel="noreferrer" variant="soft">
            {t("createAwsAccessReviewSourcePage.actions.installViaCloudFormation")}
          </ButtonAnchor>
        )}
        <Button
          type="button"
          variant="soft"
          onClick={() => copyValue(
            awsConnectorSetup.terraformSnippet,
            t("createAwsAccessReviewSourcePage.messages.copiedTerraform"),
            t("createAwsAccessReviewSourcePage.messages.copyFailed"),
          )}
        >
          {t("createAwsAccessReviewSourcePage.actions.installViaTerraform")}
        </Button>
      </div>
    );
  }

  if (provider === "GCP") {
    return (
      <Button
        type="button"
        variant="soft"
        className="shrink-0"
        onClick={() => copyValue(
          gcpConnectorSetup.terraformSnippet,
          t("createGcpAccessReviewSourcePage.messages.copiedTerraform"),
          t("createGcpAccessReviewSourcePage.messages.copyFailed"),
        )}
      >
        {t("createGcpAccessReviewSourcePage.actions.installViaTerraform")}
      </Button>
    );
  }

  if (provider === "AZURE") {
    return (
      <Button
        type="button"
        variant="soft"
        className="shrink-0"
        onClick={() => copyValue(
          azureConnectorSetup.terraformSnippet,
          t("createAzureAccessReviewSourcePage.messages.copiedTerraform"),
          t("createAzureAccessReviewSourcePage.messages.copyFailed"),
        )}
      >
        {t("createAzureAccessReviewSourcePage.actions.installViaTerraform")}
      </Button>
    );
  }

  return null;
}

function AwsWorkloadIdentityForm({
  organizationId,
  documentationUrl,
  setup,
}: {
  organizationId: string;
  documentationUrl: string | null | undefined;
  setup: ConnectVendorPageQuery["response"]["awsConnectorSetup"];
}) {
  const { t } = useTranslation("organizations/access-reviews");
  const copyValue = useCopyValue();
  const [roleArn, setRoleArn] = useState("");
  const { finish, isCreating } = useFinishWorkloadIdentity(organizationId);
  const roleArnValid = isAWSRoleARN(roleArn);
  const pageKey = "createAwsAccessReviewSourcePage";

  const onSubmit = () => {
    if (!roleArnValid) {
      return;
    }
    void finish(
      {
        organizationId,
        provider: "AWS",
        awsRoleArn: roleArn.trim(),
      },
      {
        create: t(`${pageKey}.errors.create`),
        disconnected: t(`${pageKey}.errors.disconnected`),
        delete: t(`${pageKey}.errors.delete`),
        errorTitle: t(`${pageKey}.messages.error`),
      },
    );
  };

  return (
    <form className="flex flex-col gap-4" onSubmit={(event) => { event.preventDefault(); void onSubmit(); }}>
      <Field
        label={t("createAwsAccessReviewSourcePage.fields.roleArn")}
        required
        error={roleArn.trim() !== "" && !roleArnValid ? t("createAwsAccessReviewSourcePage.errors.roleArn") : undefined}
      >
        <TextField value={roleArn} onChange={event => setRoleArn(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("createAwsAccessReviewSourcePage.fields.roleArnHelp")}</Text>
      <SetupValues rows={setupRows(t, pageKey, setup, copyValue)} />
      <ConnectFormFooter documentationUrl={documentationUrl} disabled={!roleArnValid} loading={isCreating} />
    </form>
  );
}

function GcpWorkloadIdentityForm({
  organizationId,
  documentationUrl,
  setup,
}: {
  organizationId: string;
  documentationUrl: string | null | undefined;
  setup: ConnectVendorPageQuery["response"]["gcpConnectorSetup"];
}) {
  const { t } = useTranslation("organizations/access-reviews");
  const copyValue = useCopyValue();
  const [providerResource, setProviderResource] = useState("");
  const [serviceAccountEmail, setServiceAccountEmail] = useState("");
  const { finish, isCreating } = useFinishWorkloadIdentity(organizationId);
  const providerValid = isGCPWorkloadIdentityProvider(providerResource);
  const emailValid = isGCPServiceAccountEmail(serviceAccountEmail);
  const pageKey = "createGcpAccessReviewSourcePage";

  const onSubmit = () => {
    if (!providerValid || !emailValid) {
      return;
    }
    void finish(
      {
        organizationId,
        provider: "GCP",
        gcpWorkloadIdentityProvider: providerResource.trim(),
        gcpServiceAccountEmail: serviceAccountEmail.trim(),
      },
      {
        create: t(`${pageKey}.errors.create`),
        disconnected: t(`${pageKey}.errors.disconnected`),
        delete: t(`${pageKey}.errors.delete`),
        errorTitle: t(`${pageKey}.messages.error`),
      },
    );
  };

  return (
    <form className="flex flex-col gap-4" onSubmit={(event) => { event.preventDefault(); void onSubmit(); }}>
      <Field
        label={t("createGcpAccessReviewSourcePage.fields.workloadIdentityProvider")}
        required
        error={providerResource.trim() !== "" && !providerValid
          ? t("createGcpAccessReviewSourcePage.errors.workloadIdentityProvider")
          : undefined}
      >
        <TextField value={providerResource} onChange={event => setProviderResource(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("createGcpAccessReviewSourcePage.fields.workloadIdentityProviderHelp")}</Text>
      <Field
        label={t("createGcpAccessReviewSourcePage.fields.serviceAccountEmail")}
        required
        error={serviceAccountEmail.trim() !== "" && !emailValid
          ? t("createGcpAccessReviewSourcePage.errors.serviceAccountEmail")
          : undefined}
      >
        <TextField value={serviceAccountEmail} onChange={event => setServiceAccountEmail(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("createGcpAccessReviewSourcePage.fields.serviceAccountEmailHelp")}</Text>
      <SetupValues rows={setupRows(t, pageKey, setup, copyValue)} />
      <ConnectFormFooter documentationUrl={documentationUrl} disabled={!providerValid || !emailValid} loading={isCreating} />
    </form>
  );
}

function AzureWorkloadIdentityForm({
  organizationId,
  documentationUrl,
  setup,
}: {
  organizationId: string;
  documentationUrl: string | null | undefined;
  setup: ConnectVendorPageQuery["response"]["azureConnectorSetup"];
}) {
  const { t } = useTranslation("organizations/access-reviews");
  const copyValue = useCopyValue();
  const [tenantId, setTenantId] = useState("");
  const [clientId, setClientId] = useState("");
  const [subscriptionId, setSubscriptionId] = useState("");
  const [environment, setEnvironment] = useState<AzureEnvironment>("AZURE_PUBLIC");
  const { finish, isCreating } = useFinishWorkloadIdentity(organizationId);
  const tenantValid = isAzureGUID(tenantId);
  const clientValid = isAzureGUID(clientId);
  const subscriptionValid = isAzureGUID(subscriptionId);
  const formValid = tenantValid && clientValid && subscriptionValid;
  const pageKey = "createAzureAccessReviewSourcePage";

  const onSubmit = () => {
    if (!formValid) {
      return;
    }
    void finish(
      {
        organizationId,
        provider: "AZURE",
        azureTenantId: tenantId.trim(),
        azureClientId: clientId.trim(),
        azureSubscriptionId: subscriptionId.trim(),
        azureEnvironment: environment,
      },
      {
        create: t(`${pageKey}.errors.create`),
        disconnected: t(`${pageKey}.errors.disconnected`),
        delete: t(`${pageKey}.errors.delete`),
        errorTitle: t(`${pageKey}.messages.error`),
      },
    );
  };

  return (
    <form className="flex flex-col gap-4" onSubmit={(event) => { event.preventDefault(); void onSubmit(); }}>
      <Field
        label={t("createAzureAccessReviewSourcePage.fields.tenantId")}
        required
        error={tenantId.trim() !== "" && !tenantValid ? t("createAzureAccessReviewSourcePage.errors.tenantId") : undefined}
      >
        <TextField value={tenantId} onChange={event => setTenantId(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("createAzureAccessReviewSourcePage.fields.tenantIdHelp")}</Text>
      <Field
        label={t("createAzureAccessReviewSourcePage.fields.clientId")}
        required
        error={clientId.trim() !== "" && !clientValid ? t("createAzureAccessReviewSourcePage.errors.clientId") : undefined}
      >
        <TextField value={clientId} onChange={event => setClientId(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("createAzureAccessReviewSourcePage.fields.clientIdHelp")}</Text>
      <Field
        label={t("createAzureAccessReviewSourcePage.fields.subscriptionId")}
        required
        error={subscriptionId.trim() !== "" && !subscriptionValid ? t("createAzureAccessReviewSourcePage.errors.subscriptionId") : undefined}
      >
        <TextField value={subscriptionId} onChange={event => setSubscriptionId(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("createAzureAccessReviewSourcePage.fields.subscriptionIdHelp")}</Text>
      <Field label={t("createAzureAccessReviewSourcePage.fields.environment")}>
        <Select
          value={environment}
          onValueChange={(value: string | null) => {
            if (value != null && (azureEnvironments as ReadonlyArray<string>).includes(value)) {
              setEnvironment(value as AzureEnvironment);
            }
          }}
        >
          <SelectTrigger>
            {(value: AzureEnvironment | null) => (
              value == null
                ? null
                : t(`createAzureAccessReviewSourcePage.environments.${value}`)
            )}
          </SelectTrigger>
          <SelectPopup>
            {azureEnvironments.map(value => (
              <SelectItem key={value} value={value}>
                {t(`createAzureAccessReviewSourcePage.environments.${value}`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </Field>
      <Text size={1} color="faint">{t("createAzureAccessReviewSourcePage.fields.environmentGccNote")}</Text>
      <Text size={2} color="faint">{t("createAzureAccessReviewSourcePage.propagationHint")}</Text>
      <SetupValues rows={setupRows(t, pageKey, setup, copyValue)} />
      <ConnectFormFooter documentationUrl={documentationUrl} disabled={!formValid} loading={isCreating} />
    </form>
  );
}
