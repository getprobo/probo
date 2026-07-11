// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
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
            title: __("Connection failed"),
            description: __("Failed to connect provider. Please check your credentials and try again."),
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
          title: __("Connection failed"),
          description: __("Failed to connect provider. Please check your credentials and try again."),
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
        ? sprintf(__("Connect %s"), provider.displayName)
        : __("Connect provider")}
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          connectClientCredentialsProvider();
        }}
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {sprintf(
              __("Enter the client credentials for %s to connect it as an access source."),
              provider?.displayName ?? "",
            )}
          </p>
          <Field
            label={__("Client ID")}
            value={clientId}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setClientId(e.target.value)}
            required
            autoFocus
          />
          <Field
            label={__("Client Secret")}
            type="password"
            value={clientSecret}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setClientSecret(e.target.value)}
            required
          />
          <Field
            label={__("Token URL")}
            value={tokenUrl}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTokenUrl(e.target.value)}
            required
          />
          <Field
            label={__("Scope")}
            value={scope}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setScope(e.target.value)}
          />
          {provider?.extraSettings.map(setting =>
            setting.key === "region"
              ? (
                  <div key={setting.key} className="space-y-1.5">
                    <label className="text-sm font-medium">{__(setting.label)}</label>
                    <Select
                      value={clientCredentialsExtraValues[setting.key] ?? ""}
                      onValueChange={(val: string) =>
                        setClientCredentialsExtraValues(prev => ({
                          ...prev,
                          [setting.key]: val,
                        }))}
                      placeholder={__("Select a region")}
                    >
                      <Option value="US">United States (US)</Option>
                      <Option value="CA">Canada (CA)</Option>
                      <Option value="EU">Europe (EU)</Option>
                    </Select>
                  </div>
                )
              : (
                  <Field
                    key={setting.key}
                    label={__(setting.label)}
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
            {isConnectingClientCredentials ? __("Connecting...") : __("Connect")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
