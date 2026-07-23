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

import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ClientCredentialsConnectorDialogCreateClientCredentialsConnectorMutation } from "#/__generated__/core/ClientCredentialsConnectorDialogCreateClientCredentialsConnectorMutation.graphql";

import { useCreateAccessReviewSource } from "../_hooks/useCreateAccessReviewSource";
import {
  buildExtraFields,
  hasRequiredExtraSettings,
  mapClientCredentialsExtraSettingToField,
} from "../_lib/connectorSettings";
import type { ProviderInfo } from "../AddAccessReviewSourceDialog";

const createClientCredentialsConnectorMutation = graphql`
  mutation ClientCredentialsConnectorDialogCreateClientCredentialsConnectorMutation(
    $input: CreateClientCredentialsConnectorInput!
  ) {
    createClientCredentialsConnector(input: $input) {
      connector {
        id
        provider
      }
    }
  }
`;

type Props = {
  provider: ProviderInfo | null;
  organizationId: string;
  connectionId: string;
  onClose: () => void;
  onSuccess: () => void;
};

export function ClientCredentialsConnectorDialog({
  provider,
  organizationId,
  connectionId,
  onClose,
  onSuccess,
}: Props) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [tokenUrl, setTokenUrl] = useState("");
  const [scope, setScope] = useState("");
  const [clientCredentialsExtraValues, setClientCredentialsExtraValues] = useState<Record<string, string>>({});
  const [isConnectingClientCredentials, setIsConnectingClientCredentials] = useState(false);

  const [createClientCredentialsConnector]
    = useMutation<ClientCredentialsConnectorDialogCreateClientCredentialsConnectorMutation>(
      createClientCredentialsConnectorMutation,
    );

  const createSourceAfterConnector = useCreateAccessReviewSource({
    organizationId,
    connectionId,
    onSuccess,
  });

  // Opening is driven imperatively by the parent's active-provider state. Form
  // state is reset when the dialog closes (onClose for a dismiss, the success
  // callback otherwise), so opening only shows the dialog — no setState here.
  useEffect(() => {
    if (provider) {
      dialogRef.current?.open();
    }
  }, [provider]);

  const connectClientCredentialsProvider = () => {
    if (!provider || !clientId.trim() || !clientSecret.trim() || !tokenUrl.trim()) {
      return;
    }

    const requiredSettings = provider.extraSettings.filter(s => s.required);
    if (!hasRequiredExtraSettings(requiredSettings, clientCredentialsExtraValues)) {
      return;
    }

    setIsConnectingClientCredentials(true);

    const extraFields = buildExtraFields(
      provider,
      clientCredentialsExtraValues,
      mapClientCredentialsExtraSettingToField,
    );

    createClientCredentialsConnector({
      variables: {
        input: {
          organizationId,
          provider: provider.provider,
          clientId: clientId.trim(),
          clientSecret: clientSecret.trim(),
          tokenUrl: tokenUrl.trim(),
          scope: scope.trim() || null,
          ...extraFields,
        },
      },
      onCompleted: (response) => {
        const connector = response.createClientCredentialsConnector?.connector;
        if (!connector) {
          setIsConnectingClientCredentials(false);
          toast({
            title: t("clientCredentialsConnectorDialog.messages.connectionFailed"),
            description: t("clientCredentialsConnectorDialog.errors.connect"),
            variant: "error",
          });
          return;
        }

        createSourceAfterConnector(
          connector.id,
          provider.displayName,
          () => {
            setIsConnectingClientCredentials(false);
            setClientId("");
            setClientSecret("");
            setTokenUrl("");
            setScope("");
            setClientCredentialsExtraValues({});
            dialogRef.current?.close();
            onClose();
          },
        );
      },
      onError: () => {
        setIsConnectingClientCredentials(false);
        toast({
          title: t("clientCredentialsConnectorDialog.messages.connectionFailed"),
          description: t("clientCredentialsConnectorDialog.errors.connect"),
          variant: "error",
        });
      },
    });
  };

  const clientCredentialsExtraSettingsValid = provider
    ? hasRequiredExtraSettings(provider.extraSettings, clientCredentialsExtraValues)
    : true;

  return (
    <Dialog
      ref={dialogRef}
      onClose={() => {
        // Reset on dismiss so the next open starts fresh (the imperative
        // close() on success does not fire onClose, so success resets inline).
        setClientId("");
        setClientSecret("");
        setTokenUrl("");
        setScope("");
        setClientCredentialsExtraValues({});
        setIsConnectingClientCredentials(false);
        onClose();
      }}
      title={provider
        ? t("clientCredentialsConnectorDialog.titleWithProvider", {
            provider: provider.displayName,
          })
        : t("clientCredentialsConnectorDialog.title")}
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          connectClientCredentialsProvider();
        }}
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {t("clientCredentialsConnectorDialog.description", {
              provider: provider?.displayName ?? "",
            })}
          </p>
          <Field
            label={t("clientCredentialsConnectorDialog.fields.clientId")}
            value={clientId}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setClientId(e.target.value)}
            required
            autoFocus
          />
          <Field
            label={t("clientCredentialsConnectorDialog.fields.clientSecret")}
            type="password"
            value={clientSecret}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setClientSecret(e.target.value)}
            required
          />
          <Field
            label={t("clientCredentialsConnectorDialog.fields.tokenUrl")}
            value={tokenUrl}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTokenUrl(e.target.value)}
            required
          />
          <Field
            label={t("clientCredentialsConnectorDialog.fields.scope")}
            value={scope}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setScope(e.target.value)}
          />
          {provider?.extraSettings.map(setting =>
            setting.key === "region"
              ? (
                  <div key={setting.key} className="space-y-1.5">
                    <label className="text-sm font-medium">{setting.label}</label>
                    <Select
                      value={clientCredentialsExtraValues[setting.key] ?? ""}
                      onValueChange={(val: string) =>
                        setClientCredentialsExtraValues(prev => ({
                          ...prev,
                          [setting.key]: val,
                        }))}
                      placeholder={t("clientCredentialsConnectorDialog.region.placeholder")}
                    >
                      <Option value="US">
                        {t("clientCredentialsConnectorDialog.region.us")}
                      </Option>
                      <Option value="CA">
                        {t("clientCredentialsConnectorDialog.region.ca")}
                      </Option>
                      <Option value="EU">
                        {t("clientCredentialsConnectorDialog.region.eu")}
                      </Option>
                    </Select>
                  </div>
                )
              : (
                  <Field
                    key={setting.key}
                    label={setting.label}
                    value={clientCredentialsExtraValues[setting.key] ?? ""}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                      setClientCredentialsExtraValues(prev => ({
                        ...prev,
                        [setting.key]: e.target.value,
                      }))}
                    required={setting.required}
                  />
                ),
          )}
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={
              isConnectingClientCredentials
              || !clientId.trim()
              || !clientSecret.trim()
              || !tokenUrl.trim()
              || !clientCredentialsExtraSettingsValid
            }
          >
            {isConnectingClientCredentials
              ? t("clientCredentialsConnectorDialog.actions.connecting")
              : t("clientCredentialsConnectorDialog.actions.connect")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
