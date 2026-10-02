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
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { GCPConnectForm_provider$key } from "#/__generated__/core/GCPConnectForm_provider.graphql";
import type { GCPConnectFormQuery } from "#/__generated__/core/GCPConnectFormQuery.graphql";

import { isGCPServiceAccountEmail, isGCPWorkloadIdentityProvider } from "../_lib/connectorSettings";
import { useCopyValue } from "../_lib/useCopyValue";
import { useFinishWorkloadIdentity } from "../_lib/useFinishWorkloadIdentity";
import { setupRows } from "../_lib/workloadIdentitySetup";

import { ConnectForm } from "./ConnectForm";
import { TerraformInstallButton } from "./TerraformInstallButton";
import { WorkloadIdentitySetupValues } from "./WorkloadIdentitySetupValues";

const gcpConnectFormProviderFragment = graphql`
  fragment GCPConnectForm_provider on ConnectorProviderInfo {
    ...ConnectForm_provider
  }
`;

const gcpConnectFormQuery = graphql`
  query GCPConnectFormQuery($organizationId: ID!) {
    gcpConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
    }
  }
`;

const pageKey = "marketplacePage.workloadIdentity";

export function GCPConnectInstallActions({
  organizationId,
}: {
  organizationId: string;
}) {
  const data = useLazyLoadQuery<GCPConnectFormQuery>(gcpConnectFormQuery, { organizationId });

  return <TerraformInstallButton snippet={data.gcpConnectorSetup.terraformSnippet} />;
}

export function GCPConnectForm({
  organizationId,
  providerKey,
}: {
  organizationId: string;
  providerKey: GCPConnectForm_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const provider = useFragment(gcpConnectFormProviderFragment, providerKey);
  const data = useLazyLoadQuery<GCPConnectFormQuery>(gcpConnectFormQuery, { organizationId });
  const copyValue = useCopyValue();
  const finish = useFinishWorkloadIdentity(organizationId);
  const [providerResource, setProviderResource] = useState("");
  const [serviceAccountEmail, setServiceAccountEmail] = useState("");
  const providerValid = isGCPWorkloadIdentityProvider(providerResource);
  const emailValid = isGCPServiceAccountEmail(serviceAccountEmail);

  return (
    <ConnectForm
      providerKey={provider}
      canSubmit={providerValid && emailValid}
      onSubmit={({ name }) => finish(
        {
          organizationId,
          name,
          provider: "GCP",
          gcpWorkloadIdentityProvider: providerResource.trim(),
          gcpServiceAccountEmail: serviceAccountEmail.trim(),
        },
        {
          create: t(`${pageKey}.errors.gcp.create`),
          disconnected: t(`${pageKey}.errors.gcp.disconnected`),
          delete: t(`${pageKey}.errors.delete`),
          errorTitle: t(`${pageKey}.messages.error`),
        },
      )}
    >
      <Field
        label={t("marketplacePage.workloadIdentity.fields.workloadIdentityProvider")}
        required
        error={providerResource.trim() !== "" && !providerValid
          ? t("marketplacePage.workloadIdentity.errors.workloadIdentityProvider")
          : undefined}
      >
        <TextField value={providerResource} onChange={event => setProviderResource(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.workloadIdentityProviderHelp")}</Text>
      <Field
        label={t("marketplacePage.workloadIdentity.fields.serviceAccountEmail")}
        required
        error={serviceAccountEmail.trim() !== "" && !emailValid
          ? t("marketplacePage.workloadIdentity.errors.serviceAccountEmail")
          : undefined}
      >
        <TextField value={serviceAccountEmail} onChange={event => setServiceAccountEmail(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.serviceAccountEmailHelp")}</Text>
      <WorkloadIdentitySetupValues rows={setupRows(t, pageKey, data.gcpConnectorSetup, copyValue)} />
    </ConnectForm>
  );
}
