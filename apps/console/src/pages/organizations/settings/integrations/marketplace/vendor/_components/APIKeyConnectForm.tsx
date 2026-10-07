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
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { APIKeyConnectForm_provider$key } from "#/__generated__/core/APIKeyConnectForm_provider.graphql";
import type { APIKeyConnectFormCreateMutation } from "#/__generated__/core/APIKeyConnectFormCreateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { integrationListPath } from "../../../_lib/integrationPath";
import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapAPIKeyExtraSettingToField,
} from "../_lib/extraSettings";

import { APIKeyExtraFields } from "./APIKeyExtraFields";
import { ConnectForm } from "./ConnectForm";
import { isPostHogDeploymentSelected } from "./PostHogDeploymentField";

const apiKeyConnectFormFragment = graphql`
  fragment APIKeyConnectForm_provider on ConnectorProviderInfo {
    provider
    apiKeyManaged
    apiKeyExtraSettings {
      key
      required
    }
    ...APIKeyExtraFields_provider
    ...ConnectForm_provider
  }
`;

const createAPIKeyConnectorMutation = graphql`
  mutation APIKeyConnectFormCreateMutation($input: CreateAPIKeyConnectorInput!) {
    createAPIKeyConnector(input: $input) {
      connector {
        id
      }
    }
  }
`;

export function APIKeyConnectForm({
  providerKey,
}: {
  providerKey: APIKeyConnectForm_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const provider = useFragment(apiKeyConnectFormFragment, providerKey);
  const extraSettings = provider.apiKeyExtraSettings.map(setting => ({
    key: setting.key,
    required: setting.required,
  }));
  const [apiKey, setApiKey] = useState("");
  const [extras, setExtras] = useState<Record<string, string>>({});
  const [createAPIKeyConnector] = useMutation<APIKeyConnectFormCreateMutation>(
    createAPIKeyConnectorMutation,
  );
  const postHogValid = provider.provider !== "POSTHOG" || isPostHogDeploymentSelected(extras);
  const extrasValid = hasRequiredExtraSettings(extraSettings, extras);
  const canSubmit = (provider.apiKeyManaged || apiKey.trim() !== "") && extrasValid && postHogValid;

  return (
    <ConnectForm
      providerKey={provider}
      canSubmit={canSubmit}
      onSubmit={async ({ name }) => {
        await createAPIKeyConnector({
          variables: {
            input: {
              organizationId,
              name,
              provider: provider.provider,
              apiKey: provider.apiKeyManaged ? null : apiKey.trim(),
              ...buildExtraFields(
                provider.provider,
                extraSettings,
                extras,
                mapAPIKeyExtraSettingToField,
              ),
            },
          },
        }, { errorToast: t("marketplacePage.connectFailed") });
        return { to: integrationListPath(organizationId) };
      }}
    >
      {!provider.apiKeyManaged && (
        <Field required label={t("marketplacePage.fields.apiKey")}>
          <TextField
            name="apiKey"
            type="password"
            required
            value={apiKey}
            autoComplete="off"
            onValueChange={setApiKey}
          />
        </Field>
      )}
      <APIKeyExtraFields
        providerKey={provider}
        extras={extras}
        onChange={setExtras}
      />
    </ConnectForm>
  );
}
