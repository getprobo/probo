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

import { PlusIcon } from "@phosphor-icons/react";
import { iconButton } from "@probo/ui/src/v2/IconButton/variants";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ConnectorConnectMore_organization$key } from "#/__generated__/core/ConnectorConnectMore_organization.graphql";
import type { ConnectorConnectMore_provider$key } from "#/__generated__/core/ConnectorConnectMore_provider.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { connectMethods } from "../_lib/connectMethods";
import { connectVendorPath } from "../_lib/integrationPath";

const connectorConnectMoreOrganizationFragment = graphql`
  fragment ConnectorConnectMore_organization on Organization {
    canCreate: permission(action: "core:connector:create")
  }
`;

const connectorConnectMoreFragment = graphql`
  fragment ConnectorConnectMore_provider on ConnectorProviderInfo {
    provider
    configuredProtocols
    apiKeySupported
    apiKeyManaged
    clientCredentialsSupported
    workloadIdentitySupported
    installSupported
  }
`;

interface ConnectorConnectMoreProps {
  organizationKey: ConnectorConnectMore_organization$key;
  providerKey: ConnectorConnectMore_provider$key;
}

export function ConnectorConnectMore({
  organizationKey,
  providerKey,
}: ConnectorConnectMoreProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const organization = useFragment(connectorConnectMoreOrganizationFragment, organizationKey);
  const provider = useFragment(connectorConnectMoreFragment, providerKey);
  const methods = connectMethods({
    configuredProtocols: provider.configuredProtocols,
    apiKeySupported: provider.apiKeySupported,
    apiKeyManaged: provider.apiKeyManaged,
    clientCredentialsSupported: provider.clientCredentialsSupported,
    workloadIdentitySupported: provider.workloadIdentitySupported,
    installSupported: provider.installSupported,
  });

  if (!organization.canCreate || methods.length === 0) {
    return null;
  }

  return (
    <Link
      to={connectVendorPath(organizationId, provider.provider)}
      aria-label={t("listPage.actions.connectMore")}
      className={iconButton({ variant: "ghost", color: "neutral", size: 1 })}
    >
      <PlusIcon />
    </Link>
  );
}
