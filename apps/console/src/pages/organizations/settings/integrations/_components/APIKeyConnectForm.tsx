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
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { APIKeyConnectForm_provider$key } from "#/__generated__/core/APIKeyConnectForm_provider.graphql";
import type { APIKeyConnectFormCreateMutation } from "#/__generated__/core/APIKeyConnectFormCreateMutation.graphql";
import type { APIKeyConnectFormExtraFields_provider$key } from "#/__generated__/core/APIKeyConnectFormExtraFields_provider.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapAPIKeyExtraSettingToField,
} from "../_lib/connectorSettings";
import { integrationListPath } from "../_lib/integrationPath";
import { isPostHogDeploymentSelected } from "../_lib/postHogDeployment";

import { ConnectForm } from "./ConnectForm";
import { PostHogDeploymentField } from "./PostHogDeploymentField";

const apiKeyConnectFormFragment = graphql`
  fragment APIKeyConnectForm_provider on ConnectorProviderInfo {
    provider
    apiKeyManaged
    apiKeyExtraSettings {
      key
      label
      required
    }
    ...APIKeyConnectFormExtraFields_provider
    ...ConnectForm_provider
  }
`;

const apiKeyExtraFieldsFragment = graphql`
  fragment APIKeyConnectFormExtraFields_provider on ConnectorProviderInfo {
    provider
    apiKeyExtraSettings {
      key
      label
      required
    }
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
  organizationId,
  providerKey,
}: {
  organizationId: string;
  providerKey: APIKeyConnectForm_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const provider = useFragment(apiKeyConnectFormFragment, providerKey);
  const [apiKey, setApiKey] = useState("");
  const [extras, setExtras] = useState<Record<string, string>>({});
  const [createAPIKeyConnector] = useMutation<APIKeyConnectFormCreateMutation>(
    createAPIKeyConnectorMutation,
  );
  const postHogValid = provider.provider !== "POSTHOG" || isPostHogDeploymentSelected(extras);
  const extrasValid = hasRequiredExtraSettings(provider.apiKeyExtraSettings, extras);
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
                provider.apiKeyExtraSettings,
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
        <Field label={t("marketplacePage.fields.apiKey")} required>
          <TextField
            type="password"
            value={apiKey}
            onChange={event => setApiKey(event.target.value)}
            autoComplete="off"
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

function APIKeyExtraFields({
  providerKey,
  extras,
  onChange,
}: {
  providerKey: APIKeyConnectFormExtraFields_provider$key;
  extras: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const provider = useFragment(apiKeyExtraFieldsFragment, providerKey);
  const extraSettings = provider.apiKeyExtraSettings;

  if (provider.provider === "POSTHOG") {
    return <PostHogDeploymentField values={extras} onChange={onChange} />;
  }

  if (provider.provider === "SEGMENT") {
    return (
      <Field label={t("marketplacePage.fields.region")} required>
        <Select
          value={extras.region ?? null}
          onValueChange={(value: string | null) => {
            onChange({ ...extras, region: value ?? "" });
          }}
        >
          <SelectTrigger placeholder={t("marketplacePage.fields.regionPlaceholder")}>
            {(value: string | null) => {
              if (value === "US") {
                return t("marketplacePage.fields.regionUnitedStates");
              }
              if (value === "EU") {
                return t("marketplacePage.fields.regionEurope");
              }
              return null;
            }}
          </SelectTrigger>
          <SelectPopup>
            <SelectItem value="US">{t("marketplacePage.fields.regionUnitedStates")}</SelectItem>
            <SelectItem value="EU">{t("marketplacePage.fields.regionEurope")}</SelectItem>
          </SelectPopup>
        </Select>
      </Field>
    );
  }

  return extraSettings.map(setting => (
    <Field key={setting.key} label={setting.label} required={setting.required}>
      <TextField
        value={extras[setting.key] ?? ""}
        onChange={event => onChange({ ...extras, [setting.key]: event.target.value })}
      />
    </Field>
  ));
}
