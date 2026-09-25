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

import { ThirdPartyLogo } from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { graphql, useFragment } from "react-relay";

import type { AccessReviewSourceProviderListItem_provider$key } from "#/__generated__/core/AccessReviewSourceProviderListItem_provider.graphql";

import { connectorCard } from "#/pages/organizations/settings/integrations/variants";
import { connectVendorPath } from "#/pages/organizations/settings/integrations/_lib/integrationPath";
import { ConnectorDocumentationLink } from "../../dialogs/_components/ConnectorDocumentationLink";
import { connectMethods } from "../_lib/connectMethods";

export const accessReviewSourceProviderListItemFragment = graphql`
  fragment AccessReviewSourceProviderListItem_provider on ConnectorProviderInfo {
    provider
    displayName
    documentationUrl
    configuredProtocols
    apiKeySupported
    apiKeyManaged
    clientCredentialsSupported
    workloadIdentitySupported
    installSupported
    oauth2Scopes
    ...APIKeyConnectorDialog_provider
    ...ClientCredentialsConnectorDialog_provider
    ...OAuthExtraDialog_provider
  }
`;

interface AccessReviewSourceProviderListItemProps {
  providerKey: AccessReviewSourceProviderListItem_provider$key;
  organizationId: string;
}

export function AccessReviewSourceProviderListItem({
  providerKey,
  organizationId,
}: AccessReviewSourceProviderListItemProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const navigate = useNavigate();
  const provider = useFragment(
    accessReviewSourceProviderListItemFragment,
    providerKey,
  );
  const { card, identity, title } = connectorCard();
  const methods = connectMethods({
    configuredProtocols: provider.configuredProtocols,
    apiKeySupported: provider.apiKeySupported,
    apiKeyManaged: provider.apiKeyManaged,
    clientCredentialsSupported: provider.clientCredentialsSupported,
    workloadIdentitySupported: provider.workloadIdentitySupported,
    installSupported: provider.installSupported,
  });
  const connectLabel = t("marketplacePage.connect");

  return (
    <Card
      variant="soft"
      size={2}
      padding="none"
      interactive={methods.length > 0}
      className={card()}
    >
      <div className="pointer-events-none flex items-center gap-4 px-4 py-4">
        <ThirdPartyLogo
          thirdParty={provider.provider}
          className="size-8 shrink-0"
        />
        <div className={identity()}>
          <Heading level={2} size={3} weight="medium" highContrast className={title()}>
            {provider.displayName}
          </Heading>
          <div className="pointer-events-auto relative z-1">
            <ConnectorDocumentationLink url={provider.documentationUrl} />
          </div>
        </div>
      </div>
      {methods.length > 0 && (
        <button
          type="button"
          className="absolute inset-0 z-0 cursor-pointer"
          aria-label={connectLabel}
          onClick={() => {
            void navigate(connectVendorPath(organizationId, provider.provider));
          }}
        />
      )}
    </Card>
  );
}
