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
import { graphql, usePreloadedQuery, type PreloadedQuery } from "react-relay";

import type { AzureConnectFormQuery } from "#/__generated__/core/AzureConnectFormQuery.graphql";
import type { AzureEnvironment } from "#/__generated__/core/workloadIdentityCreateMutation.graphql";

import { isAzureGUID } from "../_lib/connectorSettings";

import { ConnectForm } from "./ConnectForm";
import { SetupValues, setupRows, TerraformInstallButton, useCopyValue, useFinishWorkloadIdentity } from "./workloadIdentity";

const azureEnvironments = [
  "AZURE_PUBLIC",
  "AZURE_GOVERNMENT",
  "AZURE_GOVERNMENT_DOD",
  "AZURE_CHINA",
] as const satisfies ReadonlyArray<AzureEnvironment>;

export const azureConnectFormQuery = graphql`
  query AzureConnectFormQuery($organizationId: ID!) {
    azureConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
    }
  }
`;

const pageKey = "marketplacePage.workloadIdentity";

export function AzureConnectInstallActions({
  queryRef,
}: {
  queryRef: PreloadedQuery<AzureConnectFormQuery>;
}) {
  const data = usePreloadedQuery<AzureConnectFormQuery>(azureConnectFormQuery, queryRef);

  return <TerraformInstallButton snippet={data.azureConnectorSetup.terraformSnippet} />;
}

export function AzureConnectForm({
  organizationId,
  documentationUrl,
  queryRef,
}: {
  organizationId: string;
  documentationUrl: string | null | undefined;
  queryRef: PreloadedQuery<AzureConnectFormQuery>;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const data = usePreloadedQuery<AzureConnectFormQuery>(azureConnectFormQuery, queryRef);
  const copyValue = useCopyValue();
  const finish = useFinishWorkloadIdentity(organizationId);
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
      documentationUrl={documentationUrl}
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
        label={t("marketplacePage.workloadIdentity.fields.tenantId")}
        required
        error={tenantId.trim() !== "" && !tenantValid ? t("marketplacePage.workloadIdentity.errors.tenantId") : undefined}
      >
        <TextField value={tenantId} onChange={event => setTenantId(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.tenantIdHelp")}</Text>
      <Field
        label={t("marketplacePage.workloadIdentity.fields.clientId")}
        required
        error={clientId.trim() !== "" && !clientValid ? t("marketplacePage.workloadIdentity.errors.clientId") : undefined}
      >
        <TextField value={clientId} onChange={event => setClientId(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.clientIdHelp")}</Text>
      <Field
        label={t("marketplacePage.workloadIdentity.fields.subscriptionId")}
        required
        error={subscriptionId.trim() !== "" && !subscriptionValid
          ? t("marketplacePage.workloadIdentity.errors.subscriptionId")
          : undefined}
      >
        <TextField value={subscriptionId} onChange={event => setSubscriptionId(event.target.value)} />
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
      <SetupValues rows={setupRows(t, pageKey, data.azureConnectorSetup, copyValue)} />
    </ConnectForm>
  );
}
