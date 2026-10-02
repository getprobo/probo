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

import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { AWSConnectForm_provider$key } from "#/__generated__/core/AWSConnectForm_provider.graphql";
import type { AWSConnectFormQuery } from "#/__generated__/core/AWSConnectFormQuery.graphql";

import { isAWSRoleARN } from "../_lib/connectorSettings";
import { useCopyValue } from "../_lib/useCopyValue";
import { useFinishWorkloadIdentity } from "../_lib/useFinishWorkloadIdentity";
import { setupRows } from "../_lib/workloadIdentitySetup";

import { ConnectForm } from "./ConnectForm";
import { TerraformInstallButton } from "./TerraformInstallButton";
import { WorkloadIdentitySetupValues } from "./WorkloadIdentitySetupValues";

const awsConnectFormProviderFragment = graphql`
  fragment AWSConnectForm_provider on ConnectorProviderInfo {
    ...ConnectForm_provider
  }
`;

const awsConnectFormQuery = graphql`
  query AWSConnectFormQuery($organizationId: ID!) {
    awsConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
      cloudFormationQuickCreateURL
    }
  }
`;

const pageKey = "marketplacePage.workloadIdentity";

export function AWSConnectInstallActions({
  organizationId,
}: {
  organizationId: string;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const data = useLazyLoadQuery<AWSConnectFormQuery>(awsConnectFormQuery, { organizationId });
  const setup = data.awsConnectorSetup;

  return (
    <div className="flex shrink-0 flex-wrap justify-end gap-2">
      {setup.cloudFormationQuickCreateURL != null && (
        <ButtonAnchor href={setup.cloudFormationQuickCreateURL} target="_blank" rel="noreferrer" variant="soft">
          {t("marketplacePage.workloadIdentity.actions.installViaCloudFormation")}
        </ButtonAnchor>
      )}
      <TerraformInstallButton snippet={setup.terraformSnippet} />
    </div>
  );
}

export function AWSConnectForm({
  organizationId,
  providerKey,
}: {
  organizationId: string;
  providerKey: AWSConnectForm_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const provider = useFragment(awsConnectFormProviderFragment, providerKey);
  const data = useLazyLoadQuery<AWSConnectFormQuery>(awsConnectFormQuery, { organizationId });
  const copyValue = useCopyValue();
  const finish = useFinishWorkloadIdentity(organizationId);
  const [roleArn, setRoleArn] = useState("");
  const roleArnValid = isAWSRoleARN(roleArn);
  const setup = data.awsConnectorSetup;

  return (
    <ConnectForm
      providerKey={provider}
      canSubmit={roleArnValid}
      onSubmit={({ name }) => finish(
        {
          organizationId,
          name,
          provider: "AWS",
          awsRoleArn: roleArn.trim(),
        },
        {
          create: t(`${pageKey}.errors.aws.create`),
          disconnected: t(`${pageKey}.errors.aws.disconnected`),
          delete: t(`${pageKey}.errors.delete`),
          errorTitle: t(`${pageKey}.messages.error`),
        },
      )}
    >
      <Field
        label={t("marketplacePage.workloadIdentity.fields.roleArn")}
        required
        error={roleArn.trim() !== "" && !roleArnValid ? t("marketplacePage.workloadIdentity.errors.roleArn") : undefined}
      >
        <TextField value={roleArn} onChange={event => setRoleArn(event.target.value)} />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.roleArnHelp")}</Text>
      <WorkloadIdentitySetupValues rows={setupRows(t, pageKey, setup, copyValue)} />
    </ConnectForm>
  );
}
