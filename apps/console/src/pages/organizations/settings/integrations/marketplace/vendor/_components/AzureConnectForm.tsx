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
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AzureConnectForm_provider$key } from "#/__generated__/core/AzureConnectForm_provider.graphql";
import type { AzureConnectForm_setup$key } from "#/__generated__/core/AzureConnectForm_setup.graphql";
import type { AzureEnvironment } from "#/__generated__/core/useFinishWorkloadIdentityCreateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useCopyValue } from "../_lib/useCopyValue";
import { useFinishWorkloadIdentity } from "../_lib/useFinishWorkloadIdentity";
import { setupRows } from "../_lib/workloadIdentitySetup";

import { ConnectForm } from "./ConnectForm";
import { WorkloadIdentitySetupValues } from "./WorkloadIdentitySetupValues";

const azureEnvironments = [
  "AZURE_PUBLIC",
  "AZURE_GOVERNMENT",
  "AZURE_GOVERNMENT_DOD",
  "AZURE_CHINA",
] as const satisfies ReadonlyArray<AzureEnvironment>;

const azureConnectFormProviderFragment = graphql`
  fragment AzureConnectForm_provider on ConnectorProviderInfo {
    ...ConnectForm_provider
  }
`;

const azureConnectFormSetupFragment = graphql`
  fragment AzureConnectForm_setup on AzureConnectorSetup {
    issuer
    audience
    subject
  }
`;

const pageKey = "marketplacePage.workloadIdentity";

export function AzureConnectForm({
  providerKey,
  setupKey,
}: {
  providerKey: AzureConnectForm_provider$key;
  setupKey: AzureConnectForm_setup$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const provider = useFragment(azureConnectFormProviderFragment, providerKey);
  const setup = useFragment(azureConnectFormSetupFragment, setupKey);
  const copyValue = useCopyValue();
  const finish = useFinishWorkloadIdentity();
  const [tenantId, setTenantId] = useState("");
  const [clientId, setClientId] = useState("");
  const [subscriptionId, setSubscriptionId] = useState("");
  const [environment, setEnvironment] = useState<AzureEnvironment>("AZURE_PUBLIC");
  const tenantValid = isAzureGUID(tenantId);
  const clientValid = isAzureGUID(clientId);
  const subscriptionValid = isAzureGUID(subscriptionId);
  const formValid = tenantValid && clientValid && subscriptionValid;

  return (
    <ConnectForm
      providerKey={provider}
      canSubmit={formValid}
      onSubmit={({ name }) => finish(
        {
          organizationId,
          name,
          provider: "AZURE",
          azureTenantId: tenantId.trim(),
          azureClientId: clientId.trim(),
          azureSubscriptionId: subscriptionId.trim(),
          azureEnvironment: environment,
        },
        {
          create: t(`${pageKey}.errors.azure.create`),
          disconnected: t(`${pageKey}.errors.azure.disconnected`),
          delete: t(`${pageKey}.errors.delete`),
          errorTitle: t(`${pageKey}.messages.error`),
        },
      )}
    >
      <Field
        required
        label={t("marketplacePage.workloadIdentity.fields.tenantId")}
        error={tenantId.trim() !== "" && !tenantValid ? t("marketplacePage.workloadIdentity.errors.tenantId") : undefined}
      >
        <TextField name="tenantId" required value={tenantId} onValueChange={setTenantId} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.tenantIdHelp")}</Text>
      <Field
        required
        label={t("marketplacePage.workloadIdentity.fields.clientId")}
        error={clientId.trim() !== "" && !clientValid ? t("marketplacePage.workloadIdentity.errors.clientId") : undefined}
      >
        <TextField name="clientId" required value={clientId} onValueChange={setClientId} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.clientIdHelp")}</Text>
      <Field
        required
        label={t("marketplacePage.workloadIdentity.fields.subscriptionId")}
        error={subscriptionId.trim() !== "" && !subscriptionValid
          ? t("marketplacePage.workloadIdentity.errors.subscriptionId")
          : undefined}
      >
        <TextField
          name="subscriptionId"
          required
          value={subscriptionId}
          onValueChange={setSubscriptionId}
        />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.subscriptionIdHelp")}</Text>
      <Field label={t("marketplacePage.workloadIdentity.fields.environment")}>
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
                : t(`marketplacePage.workloadIdentity.environments.${value}`)
            )}
          </SelectTrigger>
          <SelectPopup>
            {azureEnvironments.map(value => (
              <SelectItem key={value} value={value}>
                {t(`marketplacePage.workloadIdentity.environments.${value}`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.environmentGccNote")}</Text>
      <Text size={2} color="faint">{t("marketplacePage.workloadIdentity.propagationHint")}</Text>
      <WorkloadIdentitySetupValues
        rows={setupRows(t, pageKey, {
          issuer: setup.issuer,
          audience: setup.audience,
          subject: setup.subject,
        }, copyValue)}
      />
    </ConnectForm>
  );
}

const AZURE_GUID_PATTERN
  = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

const AZURE_NIL_GUID = "00000000-0000-0000-0000-000000000000";

function isAzureGUID(value: string): boolean {
  const trimmed = value.trim();
  if (trimmed === "" || trimmed === AZURE_NIL_GUID) {
    return false;
  }

  return AZURE_GUID_PATTERN.test(trimmed);
}
