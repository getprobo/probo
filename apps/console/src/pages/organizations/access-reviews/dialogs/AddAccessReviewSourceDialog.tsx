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

import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Input,
  ThirdPartyLogo,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode, useMemo, useState } from "react";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { AddAccessReviewSourceDialogConnectorProviderInfoFragment$data } from "#/__generated__/core/AddAccessReviewSourceDialogConnectorProviderInfoFragment.graphql";

import { APIKeyConnectorDialog } from "./_components/APIKeyConnectorDialog";
import { ClientCredentialsConnectorDialog } from "./_components/ClientCredentialsConnectorDialog";
import {
  DatadogConnectDialog,
  ZendeskConnectDialog,
} from "./_components/OAuthExtraDialog";
import { connectOAuthProvider } from "./_lib/connectorSettings";

export const addAccessReviewSourceDialogConnectorProviderInfoFragment = graphql`
  fragment AddAccessReviewSourceDialogConnectorProviderInfoFragment on ConnectorProviderInfo @relay(plural: true) {
    provider
    displayName
    oauthConfigured
    apiKeySupported
    apiKeyManaged
    clientCredentialsSupported
    oauth2Scopes
    extraSettings {
      key
      label
      required
    }
  }
`;

export type ProviderInfo = AddAccessReviewSourceDialogConnectorProviderInfoFragment$data[number];

type Props = {
  children: ReactNode;
  organizationId: string;
  connectionId: string;
  providerInfos: ReadonlyArray<ProviderInfo>;
  existingSourceProviders: ReadonlyArray<string>;
};

export function AddAccessReviewSourceDialog({
  children,
  organizationId,
  connectionId,
  providerInfos,
  existingSourceProviders,
}: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  const [searchQuery, setSearchQuery] = useState("");

  const [activeAPIKeyProvider, setActiveAPIKeyProvider] = useState<ProviderInfo | null>(null);
  const [activeClientCredsProvider, setActiveClientCredsProvider] = useState<ProviderInfo | null>(null);
  const [activeDatadogProvider, setActiveDatadogProvider] = useState<ProviderInfo | null>(null);
  const [activeZendeskProvider, setActiveZendeskProvider] = useState<ProviderInfo | null>(null);

  const filteredProviders = useMemo(() => {
    const sorted = [...providerInfos].sort((a, b) =>
      a.displayName.localeCompare(b.displayName),
    );
    if (!searchQuery.trim()) return sorted;
    const q = searchQuery.toLowerCase();
    return sorted.filter(
      info => info.displayName.toLowerCase().includes(q),
    );
  }, [providerInfos, searchQuery]);

  const connectedProviders = useMemo(
    () => new Set(existingSourceProviders),
    [existingSourceProviders],
  );

  const renderProviderCard = (info: ProviderInfo) => {
    const isConnected = connectedProviders.has(info.provider);

    const hasSecondaryOptions = info.oauthConfigured
      && (info.apiKeySupported || info.clientCredentialsSupported);

    const renderPrimaryButton = () => {
      if (info.oauthConfigured) {
        return (
          <Button
            variant="secondary"
            onClick={() => {
              if (info.provider === "DATADOG") {
                setActiveDatadogProvider(info);
              } else if (info.provider === "ZENDESK") {
                setActiveZendeskProvider(info);
              } else {
                connectOAuthProvider(organizationId, info);
              }
            }}
          >
            {__("Connect")}
          </Button>
        );
      }
      if (info.apiKeySupported || info.apiKeyManaged) {
        return (
          <Button
            variant="secondary"
            onClick={() => setActiveAPIKeyProvider(info)}
          >
            {info.apiKeyManaged ? __("Connect") : __("API Key")}
          </Button>
        );
      }
      if (info.clientCredentialsSupported) {
        return (
          <Button
            variant="secondary"
            onClick={() => setActiveClientCredsProvider(info)}
          >
            {__("Client Credentials")}
          </Button>
        );
      }
      return null;
    };

    return (
      <Card key={info.provider} padded className="flex items-center gap-3">
        <ThirdPartyLogo thirdParty={info.provider} tint className="size-6 shrink-0" />
        <div className="mr-auto">
          <h3 className="font-medium">{info.displayName}</h3>
        </div>
        {isConnected
          ? (
              <Badge variant="success" size="md">
                {__("Connected")}
              </Badge>
            )
          : (
              <div className="flex items-center gap-2">
                {renderPrimaryButton()}
                {hasSecondaryOptions && (
                  <ActionDropdown variant="secondary">
                    {info.apiKeySupported && (
                      <DropdownItem
                        onSelect={() => setActiveAPIKeyProvider(info)}
                      >
                        {__("Connect with API Key")}
                      </DropdownItem>
                    )}
                    {info.clientCredentialsSupported && (
                      <DropdownItem
                        onSelect={() => setActiveClientCredsProvider(info)}
                      >
                        {__("Connect with Client Credentials")}
                      </DropdownItem>
                    )}
                  </ActionDropdown>
                )}
              </div>
            )}
      </Card>
    );
  };

  return (
    <>
      <Dialog
        ref={dialogRef}
        trigger={children}
        title={(
          <Breadcrumb
            items={[
              __("Access Reviews"),
              __("Add Source"),
            ]}
          />
        )}
      >
        <DialogContent padded className="space-y-4">
          <Input
            placeholder={__("Search providers...")}
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
          />

          <div className="space-y-3">
            {filteredProviders.map(info => renderProviderCard(info))}

            {(!searchQuery.trim() || "csv".includes(searchQuery.toLowerCase())) && (
              <Card padded className="flex items-center gap-3">
                <div className="mr-auto">
                  <h3 className="font-medium">{__("CSV")}</h3>
                  <p className="text-sm text-txt-secondary">
                    {__("Upload CSV data directly as an access source.")}
                  </p>
                </div>
                <Button
                  variant="secondary"
                  asChild
                  onClick={() => dialogRef.current?.close()}
                >
                  <Link to={`/organizations/${organizationId}/access-reviews/sources/new/csv`}>
                    {__("Open")}
                  </Link>
                </Button>
              </Card>
            )}
          </div>
        </DialogContent>
        <DialogFooter exitLabel={__("Close")} />
      </Dialog>

      <APIKeyConnectorDialog
        provider={activeAPIKeyProvider}
        organizationId={organizationId}
        connectionId={connectionId}
        onClose={() => setActiveAPIKeyProvider(null)}
        onSuccess={() => dialogRef.current?.close()}
      />

      <ClientCredentialsConnectorDialog
        provider={activeClientCredsProvider}
        organizationId={organizationId}
        connectionId={connectionId}
        onClose={() => setActiveClientCredsProvider(null)}
        onSuccess={() => dialogRef.current?.close()}
      />

      <DatadogConnectDialog
        provider={activeDatadogProvider}
        organizationId={organizationId}
        onClose={() => setActiveDatadogProvider(null)}
      />

      <ZendeskConnectDialog
        provider={activeZendeskProvider}
        organizationId={organizationId}
        onClose={() => setActiveZendeskProvider(null)}
      />
    </>
  );
}
