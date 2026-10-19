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

import { LightbulbIcon, PlugIcon } from "@phosphor-icons/react";
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { CardLink } from "@probo/ui/src/v2/Card/CardLink";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorProviderListItem_provider$key } from "#/__generated__/core/ConnectorProviderListItem_provider.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ConnectorDocumentationLink } from "../../_components/ConnectorDocumentationLink";
import { ConnectorProviderName } from "../../_components/ConnectorProviderName";
import { connectMethods } from "../../_lib/connectMethods";
import { connectVendorPath } from "../../_lib/integrationPath";
import { connectorCard } from "../../variants";

const connectorProviderListItemFragment = graphql`
  fragment ConnectorProviderListItem_provider on ConnectorProviderInfo {
    provider
    displayName
    configuredProtocols
    apiKeySupported
    apiKeyManaged
    clientCredentialsSupported
    workloadIdentitySupported
    installSupported
    ...ConnectorDocumentationLink_provider
  }
`;

interface ConnectorProviderListItemProps {
  providerKey: ConnectorProviderListItem_provider$key;
  credentialCount: number;
  matchesThirdParty: boolean;
}

export function ConnectorProviderListItem({
  providerKey,
  credentialCount,
  matchesThirdParty,
}: ConnectorProviderListItemProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const provider = useFragment(connectorProviderListItemFragment, providerKey);
  const { card } = connectorCard();
  const methods = connectMethods({
    configuredProtocols: provider.configuredProtocols,
    apiKeySupported: provider.apiKeySupported,
    apiKeyManaged: provider.apiKeyManaged,
    clientCredentialsSupported: provider.clientCredentialsSupported,
    workloadIdentitySupported: provider.workloadIdentitySupported,
    installSupported: provider.installSupported,
  });
  const connectLabel = t("marketplacePage.connect");
  const thirdPartyLabel = t("marketplacePage.thirdParties.iconLabel");

  return (
    <Card
      variant="soft"
      size={2}
      padding="none"
      interactive={methods.length > 0}
      className={card()}
    >
      {methods.length > 0 && (
        <CardLink
          to={connectVendorPath(organizationId, provider.provider)}
          aria-label={connectLabel}
        />
      )}
      <div className="pointer-events-none flex items-center gap-6 px-4 py-4">
        <ThirdPartyLogo
          thirdParty={provider.provider}
          className="size-8 shrink-0"
        />
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <Heading level={2} size={3} weight="medium" highContrast className="flex min-w-0 leading-tight">
            <ConnectorProviderName provider={provider.provider} name={provider.displayName} />
          </Heading>
          <div className="pointer-events-auto relative z-1">
            <ConnectorDocumentationLink providerKey={provider} />
          </div>
        </div>
        {(matchesThirdParty || credentialCount > 0) && (
          <div className="flex shrink-0 items-center gap-2">
            {matchesThirdParty && (
              <Tooltip>
                <TooltipTrigger
                  render={(
                    <span
                      tabIndex={0}
                      aria-label={thirdPartyLabel}
                      className="pointer-events-auto relative z-1 inline-flex text-amber-11"
                    >
                      <LightbulbIcon weight="duotone" className="size-5" />
                    </span>
                  )}
                />
                <TooltipPopup>{thirdPartyLabel}</TooltipPopup>
              </Tooltip>
            )}
            {credentialCount > 0 && (
              <Badge
                variant="soft"
                color="sky"
                size={1}
                iconStart={<PlugIcon />}
                aria-label={t("marketplacePage.credentialCount", { count: credentialCount })}
              >
                {credentialCount}
              </Badge>
            )}
          </div>
        )}
      </div>
    </Card>
  );
}
