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

import { Button, ThirdPartyLogo } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AccessReviewSourceProviderListItem_provider$key } from "#/__generated__/core/AccessReviewSourceProviderListItem_provider.graphql";
import type { ConnectorProtocol } from "#/__generated__/core/connectorProviderInfoFields_installableProtocols.graphql";

import { APIKeyConnectorDialog } from "../../dialogs/_components/APIKeyConnectorDialog";
import { ClientCredentialsConnectorDialog } from "../../dialogs/_components/ClientCredentialsConnectorDialog";
import { ConnectorDocumentationLink } from "../../dialogs/_components/ConnectorDocumentationLink";
import {
  DatadogConnectDialog,
  ZendeskConnectDialog,
} from "../../dialogs/_components/OAuthExtraDialog";
import {
  connectOAuthProvider,
  connectProviderProtocol,
} from "../../dialogs/_lib/connectorSettings";

import { accessReviewSourceSection } from "./variants";

const protocolActionLabelKey: Record<ConnectorProtocol, string> = {
  OAUTH2: "addAccessReviewSourceDialog.actions.connectWithOAuth",
  API_KEY: "addAccessReviewSourceDialog.actions.connectWithApiKey",
  GITHUB_APP: "addAccessReviewSourceDialog.actions.connectWithGitHubApp",
  WORKLOAD_IDENTITY:
    "addAccessReviewSourceDialog.actions.connectWithWorkloadIdentity",
};

export const accessReviewSourceProviderListItemFragment = graphql`
  fragment AccessReviewSourceProviderListItem_provider on ConnectorProviderInfo {
    provider
    displayName
    documentationUrl
    oauthConfigured
    installableProtocols
    apiKeySupported
    apiKeyManaged
    clientCredentialsSupported
    oauth2Scopes
    ...APIKeyConnectorDialog_provider
    ...ClientCredentialsConnectorDialog_provider
    ...OAuthExtraDialog_provider
  }
`;

interface AccessReviewSourceProviderListItemProps {
  providerKey: AccessReviewSourceProviderListItem_provider$key;
  organizationId: string;
  connectionId: string;
}

export function AccessReviewSourceProviderListItem({
  providerKey,
  organizationId,
  connectionId,
}: AccessReviewSourceProviderListItemProps) {
  const { t } = useTranslation();
  const provider = useFragment(
    accessReviewSourceProviderListItemFragment,
    providerKey,
  );
  const { item, content, trailing } = accessReviewSourceSection();
  const [activeDialog, setActiveDialog] = useState<
    "apiKey" | "clientCredentials" | "datadog" | "zendesk" | null
  >(null);

  // Every row renders the dialogs its provider can actually reach, so a list of
  // providers does not mount three unusable dialogs per row.
  const supportsAPIKey = provider.apiKeySupported || provider.apiKeyManaged;
  const supportsDatadogOAuth
    = provider.oauthConfigured && provider.provider === "DATADOG";
  const supportsZendeskOAuth
    = provider.oauthConfigured && provider.provider === "ZENDESK";

  const connectWithOAuth = () => {
    if (provider.provider === "DATADOG") {
      setActiveDialog("datadog");
    } else if (provider.provider === "ZENDESK") {
      setActiveDialog("zendesk");
    } else {
      connectOAuthProvider(
        organizationId,
        provider.provider,
        provider.oauth2Scopes,
      );
    }
  };

  return (
    <li className={item()}>
      <ThirdPartyLogo
        thirdParty={provider.provider}
        className="size-6 shrink-0"
      />
      <div className={content()}>
        <span className="text-sm font-medium text-txt-primary">
          {provider.displayName}
        </span>
        <ConnectorDocumentationLink url={provider.documentationUrl} />
      </div>
      <div className={trailing()}>
        {provider.oauthConfigured && (
          <Button variant="primary" onClick={connectWithOAuth}>
            {t("addAccessReviewSourceDialog.actions.connectWithOAuth")}
          </Button>
        )}
        {(provider.installableProtocols ?? []).map(protocol => (
          <Button
            key={protocol}
            variant="primary"
            onClick={() =>
              connectProviderProtocol(
                organizationId,
                provider.provider,
                protocol,
              )}
          >
            {t(protocolActionLabelKey[protocol])}
          </Button>
        ))}
        {supportsAPIKey && (
          <Button variant="primary" onClick={() => setActiveDialog("apiKey")}>
            {t("addAccessReviewSourceDialog.actions.connectWithApiKey")}
          </Button>
        )}
        {provider.clientCredentialsSupported && (
          <Button
            variant="primary"
            onClick={() => setActiveDialog("clientCredentials")}
          >
            {t(
              "addAccessReviewSourceDialog.actions.connectWithClientCredentials",
            )}
          </Button>
        )}
      </div>
      {supportsAPIKey && (
        <APIKeyConnectorDialog
          providerKey={activeDialog === "apiKey" ? provider : null}
          organizationId={organizationId}
          connectionId={connectionId}
          onClose={() => setActiveDialog(null)}
          onSuccess={() => setActiveDialog(null)}
        />
      )}
      {provider.clientCredentialsSupported && (
        <ClientCredentialsConnectorDialog
          providerKey={activeDialog === "clientCredentials" ? provider : null}
          organizationId={organizationId}
          connectionId={connectionId}
          onClose={() => setActiveDialog(null)}
          onSuccess={() => setActiveDialog(null)}
        />
      )}
      {supportsDatadogOAuth && (
        <DatadogConnectDialog
          providerKey={activeDialog === "datadog" ? provider : null}
          organizationId={organizationId}
          onClose={() => setActiveDialog(null)}
        />
      )}
      {supportsZendeskOAuth && (
        <ZendeskConnectDialog
          providerKey={activeDialog === "zendesk" ? provider : null}
          organizationId={organizationId}
          onClose={() => setActiveDialog(null)}
        />
      )}
    </li>
  );
}
