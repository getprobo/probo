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

import { useToast } from "@probo/ui";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { APIKeyConnectFormCreateMutation } from "#/__generated__/core/APIKeyConnectFormCreateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { isPostHogDeploymentSelected, PostHogDeploymentField } from "#/pages/organizations/access-reviews/dialogs/PostHogDeploymentField";
import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapAPIKeyExtraSettingToField,
} from "#/pages/organizations/access-reviews/dialogs/_lib/connectorSettings";

import { integrationListPath } from "../_lib/integrationPath";
import { ConnectFormFooter, type ConnectVendorDriver } from "./ConnectFormFooter";

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
  driver,
}: {
  organizationId: string;
  driver: ConnectVendorDriver;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { t: tRegions } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [apiKey, setApiKey] = useState("");
  const [extras, setExtras] = useState<Record<string, string>>({});
  const [isConnecting, setIsConnecting] = useState(false);
  const [createAPIKeyConnector] = useMutation<APIKeyConnectFormCreateMutation>(
    createAPIKeyConnectorMutation,
  );

  const postHogValid = driver.provider !== "POSTHOG" || isPostHogDeploymentSelected(extras);
  const extrasValid = hasRequiredExtraSettings(driver.apiKeyExtraSettings, extras);
  const canSubmit = (driver.apiKeyManaged || apiKey.trim() !== "") && extrasValid && postHogValid;

  const onSubmit = async () => {
    if (!canSubmit || isConnecting) {
      return;
    }
    setIsConnecting(true);
    try {
      await createAPIKeyConnector({
        variables: {
          input: {
            organizationId,
            provider: driver.provider,
            apiKey: driver.apiKeyManaged ? null : apiKey.trim(),
            ...buildExtraFields(
              driver.provider,
              driver.apiKeyExtraSettings,
              extras,
              mapAPIKeyExtraSettingToField,
            ),
          },
        },
      }, { errorToast: t("marketplacePage.connectFailed") });
      toast({
        title: t("marketplacePage.connected"),
        description: t("listPage.messages.connectedDescription"),
        variant: "success",
      });
      void navigate(integrationListPath(organizationId));
    } catch {
      return;
    } finally {
      setIsConnecting(false);
    }
  };

  return (
    <form
      className="flex flex-col gap-4"
      onSubmit={(event) => {
        event.preventDefault();
        void onSubmit();
      }}
    >
      {!driver.apiKeyManaged && (
        <Field label={t("marketplacePage.fields.apiKey")} required>
          <TextField
            value={apiKey}
            onChange={event => setApiKey(event.target.value)}
            autoComplete="off"
          />
        </Field>
      )}
      <APIKeyExtraFields driver={driver} extras={extras} onChange={setExtras} regionLabel={tRegions("accessReviewSource.regions.label")} />
      <ConnectFormFooter
        documentationUrl={driver.documentationUrl}
        disabled={!canSubmit}
        loading={isConnecting}
      />
    </form>
  );
}

function APIKeyExtraFields({
  driver,
  extras,
  onChange,
  regionLabel,
}: {
  driver: ConnectVendorDriver;
  extras: Record<string, string>;
  onChange: (values: Record<string, string>) => void;
  regionLabel: string;
}) {
  const { t } = useTranslation();

  if (driver.provider === "POSTHOG") {
    return <PostHogDeploymentField values={extras} onChange={onChange} />;
  }

  if (driver.provider === "SEGMENT") {
    return (
      <Field label={regionLabel} required>
        <Select
          value={extras.region ?? null}
          onValueChange={(value: string | null) => {
            onChange({ ...extras, region: value ?? "" });
          }}
        >
          <SelectTrigger placeholder={t("accessReviewSource.regions.placeholder")}>
            {(value: string | null) => {
              if (value === "US") {
                return t("accessReviewSource.regions.unitedStates");
              }
              if (value === "EU") {
                return t("accessReviewSource.regions.europe");
              }
              return null;
            }}
          </SelectTrigger>
          <SelectPopup>
            <SelectItem value="US">{t("accessReviewSource.regions.unitedStates")}</SelectItem>
            <SelectItem value="EU">{t("accessReviewSource.regions.europe")}</SelectItem>
          </SelectPopup>
        </Select>
      </Field>
    );
  }

  return driver.apiKeyExtraSettings.map(setting => (
    <Field key={setting.key} label={setting.label} required={setting.required}>
      <TextField
        value={extras[setting.key] ?? ""}
        onChange={event => onChange({ ...extras, [setting.key]: event.target.value })}
      />
    </Field>
  ));
}
