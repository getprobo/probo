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
import { graphql, useFragment } from "react-relay";

import type { GCPConnectForm_provider$key } from "#/__generated__/core/GCPConnectForm_provider.graphql";
import type { GCPConnectForm_setup$key } from "#/__generated__/core/GCPConnectForm_setup.graphql";

import { useCopyValue } from "../_lib/useCopyValue";
import { useFinishWorkloadIdentity } from "../_lib/useFinishWorkloadIdentity";
import { setupRows } from "../_lib/workloadIdentitySetup";

import { ConnectForm } from "./ConnectForm";
import { WorkloadIdentitySetupValues } from "./WorkloadIdentitySetupValues";

const gcpConnectFormProviderFragment = graphql`
  fragment GCPConnectForm_provider on ConnectorProviderInfo {
    ...ConnectForm_provider
  }
`;

const gcpConnectFormSetupFragment = graphql`
  fragment GCPConnectForm_setup on GCPConnectorSetup {
    issuer
    audience
    subject
  }
`;

const pageKey = "marketplacePage.workloadIdentity";

export function GCPConnectForm({
  organizationId,
  providerKey,
  setupKey,
}: {
  organizationId: string;
  providerKey: GCPConnectForm_provider$key;
  setupKey: GCPConnectForm_setup$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const provider = useFragment(gcpConnectFormProviderFragment, providerKey);
  const setup = useFragment(gcpConnectFormSetupFragment, setupKey);
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
        required
        label={t("marketplacePage.workloadIdentity.fields.workloadIdentityProvider")}
        error={providerResource.trim() !== "" && !providerValid
          ? t("marketplacePage.workloadIdentity.errors.workloadIdentityProvider")
          : undefined}
      >
        <TextField
          name="workloadIdentityProvider"
          required
          value={providerResource}
          onValueChange={setProviderResource}
        />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.workloadIdentityProviderHelp")}</Text>
      <Field
        required
        label={t("marketplacePage.workloadIdentity.fields.serviceAccountEmail")}
        error={serviceAccountEmail.trim() !== "" && !emailValid
          ? t("marketplacePage.workloadIdentity.errors.serviceAccountEmail")
          : undefined}
      >
        <TextField
          name="serviceAccountEmail"
          required
          value={serviceAccountEmail}
          onValueChange={setServiceAccountEmail}
        />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.serviceAccountEmailHelp")}</Text>
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

const GCP_PROVIDER_RESOURCE_PATTERN = new RegExp([
  "^(?:(?:https:)?\\/\\/iam\\.(?:googleapis\\.com|s3nsapis\\.fr)\\/)?",
  "projects\\/([1-9][0-9]*)\\/locations\\/global\\/workloadIdentityPools\\/",
  "([a-z0-9][a-z0-9-]{2,30}[a-z0-9])\\/providers\\/",
  "([a-z0-9][a-z0-9-]{2,30}[a-z0-9])\\/?$",
].join(""));

const GCP_SERVICE_ACCOUNT_EMAIL_PATTERN
  = /^[a-z][a-z0-9-]{4,28}[a-z0-9]@[a-z][a-z0-9-]{4,28}[a-z0-9](?:\.s3ns)?\.iam\.gserviceaccount\.com$/;

function isGCPWorkloadIdentityProvider(value: string): boolean {
  return GCP_PROVIDER_RESOURCE_PATTERN.test(value.trim());
}

function isGCPServiceAccountEmail(value: string): boolean {
  return GCP_SERVICE_ACCOUNT_EMAIL_PATTERN.test(value.trim());
}
