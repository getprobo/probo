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
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { ClientCredentialsConnectFormCreateMutation } from "#/__generated__/core/ClientCredentialsConnectFormCreateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapClientCredentialsExtraSettingToField,
} from "#/pages/organizations/access-reviews/dialogs/_lib/connectorSettings";

import { integrationListPath } from "../_lib/integrationPath";
import { ConnectFormFooter, type ConnectVendorDriver } from "./ConnectFormFooter";

const createClientCredentialsConnectorMutation = graphql`
  mutation ClientCredentialsConnectFormCreateMutation(
    $input: CreateClientCredentialsConnectorInput!
  ) {
    createClientCredentialsConnector(input: $input) {
      connector {
        id
      }
    }
  }
`;

export function ClientCredentialsConnectForm({
  organizationId,
  driver,
}: {
  organizationId: string;
  driver: ConnectVendorDriver;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const navigate = useNavigate();
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [tokenUrl, setTokenUrl] = useState("");
  const [scope, setScope] = useState("");
  const [extras, setExtras] = useState<Record<string, string>>({});
  const [isConnecting, setIsConnecting] = useState(false);
  const [createClientCredentialsConnector] = useMutation<ClientCredentialsConnectFormCreateMutation>(
    createClientCredentialsConnectorMutation,
  );
  const tokenReady = Boolean(driver.clientCredentialsTokenUrl) || tokenUrl.trim() !== "";
  const extrasValid = hasRequiredExtraSettings(driver.clientCredentialsExtraSettings, extras);
  const canSubmit = clientId.trim() !== "" && clientSecret.trim() !== "" && tokenReady && extrasValid;

  const onSubmit = async () => {
    if (!canSubmit || isConnecting) {
      return;
    }
    setIsConnecting(true);
    try {
      await createClientCredentialsConnector({
        variables: {
          input: {
            organizationId,
            provider: driver.provider,
            clientId: clientId.trim(),
            clientSecret: clientSecret.trim(),
            tokenUrl: driver.clientCredentialsTokenUrl ? null : tokenUrl.trim(),
            scope: scope.trim() || null,
            ...buildExtraFields(
              driver.provider,
              driver.clientCredentialsExtraSettings,
              extras,
              mapClientCredentialsExtraSettingToField,
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
      <Field label={t("marketplacePage.fields.clientId")} required>
        <TextField value={clientId} onChange={event => setClientId(event.target.value)} />
      </Field>
      <Field label={t("marketplacePage.fields.clientSecret")} required>
        <TextField
          type="password"
          value={clientSecret}
          onChange={event => setClientSecret(event.target.value)}
          autoComplete="off"
        />
      </Field>
      {driver.clientCredentialsTokenUrl == null && (
        <Field label={t("marketplacePage.fields.tokenUrl")} required>
          <TextField value={tokenUrl} onChange={event => setTokenUrl(event.target.value)} />
        </Field>
      )}
      <Field label={t("marketplacePage.fields.scope")}>
        <TextField value={scope} onChange={event => setScope(event.target.value)} />
      </Field>
      {driver.clientCredentialsExtraSettings.map(setting => (
        <Field key={setting.key} label={setting.label} required={setting.required}>
          <TextField
            value={extras[setting.key] ?? ""}
            onChange={event => setExtras({ ...extras, [setting.key]: event.target.value })}
          />
        </Field>
      ))}
      <ConnectFormFooter
        documentationUrl={driver.documentationUrl}
        disabled={!canSubmit}
        loading={isConnecting}
      />
    </form>
  );
}
