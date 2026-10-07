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

import type { AWSConnectForm_provider$key } from "#/__generated__/core/AWSConnectForm_provider.graphql";
import type { AWSConnectForm_setup$key } from "#/__generated__/core/AWSConnectForm_setup.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useCopyValue } from "../_lib/useCopyValue";
import { useFinishWorkloadIdentity } from "../_lib/useFinishWorkloadIdentity";
import { setupRows } from "../_lib/workloadIdentitySetup";

import { ConnectForm } from "./ConnectForm";
import { WorkloadIdentitySetupValues } from "./WorkloadIdentitySetupValues";

const awsConnectFormProviderFragment = graphql`
  fragment AWSConnectForm_provider on ConnectorProviderInfo {
    ...ConnectForm_provider
  }
`;

const awsConnectFormSetupFragment = graphql`
  fragment AWSConnectForm_setup on AWSConnectorSetup {
    issuer
    audience
    subject
  }
`;

const pageKey = "marketplacePage.workloadIdentity";

export function AWSConnectForm({
  providerKey,
  setupKey,
}: {
  providerKey: AWSConnectForm_provider$key;
  setupKey: AWSConnectForm_setup$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const provider = useFragment(awsConnectFormProviderFragment, providerKey);
  const setup = useFragment(awsConnectFormSetupFragment, setupKey);
  const copyValue = useCopyValue();
  const finish = useFinishWorkloadIdentity();
  const [roleArn, setRoleArn] = useState("");
  const roleArnValid = isAWSRoleARN(roleArn);

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
        required
        label={t("marketplacePage.workloadIdentity.fields.roleArn")}
        error={roleArn.trim() !== "" && !roleArnValid ? t("marketplacePage.workloadIdentity.errors.roleArn") : undefined}
      >
        <TextField
          name="roleArn"
          required
          value={roleArn}
          onValueChange={setRoleArn}
        />
      </Field>
      <Text size={1} color="faint">{t("marketplacePage.workloadIdentity.fields.roleArnHelp")}</Text>
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

// Same grammar as pkg/awsx/arn.RoleARNPattern, with the three supported
// partitions inlined so the field rejects other partitions immediately.
const AWS_IAM_ROLE_ARN_PATTERN
  = "arn:(aws-us-gov|aws-cn|aws):iam::([0-9]{12}):role(?:/[\\w+=,.@\\-]+)*/[\\w+=,.@\\-]{1,64}";

const awsIAMRoleARN = new RegExp(`^${AWS_IAM_ROLE_ARN_PATTERN}$`);

function isAWSRoleARN(value: string): boolean {
  return awsIAMRoleARN.test(value.trim());
}
