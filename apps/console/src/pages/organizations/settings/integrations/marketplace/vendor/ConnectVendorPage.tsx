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

import { usePageTitle } from "@probo/hooks";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, useParams } from "react-router";

import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import {
  connectMethodFromSlug,
  connectMethods,
  connectVendorMethodPath,
} from "../../_lib/connectMethods";
import { providerFromSlug } from "../../_lib/integrationPath";

import { ConnectVendorChooser } from "./_components/ConnectVendorChooser";
import { ConnectVendorMethod } from "./_components/ConnectVendorMethod";

// The page query spreads fragments owned by the forms and install actions.
import "./_components/AWSConnectForm";
import "./_components/AWSConnectInstallActions";
import "./_components/AzureConnectForm";
import "./_components/AzureConnectInstallActions";
import "./_components/GCPConnectForm";
import "./_components/GCPConnectInstallActions";

export const connectVendorPageQuery = graphql`
  query ConnectVendorPageQuery(
    $organizationId: ID!
    $includeAWS: Boolean!
    $includeAzure: Boolean!
    $includeGCP: Boolean!
  ) {
    connectorProviders {
      provider
      displayName
      configuredProtocols
      apiKeySupported
      apiKeyManaged
      clientCredentialsSupported
      workloadIdentitySupported
      installSupported
      ...ConnectVendorChooser_provider
      ...ConnectVendorMethod_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
      }
    }
    awsConnectorSetup(organizationId: $organizationId) @include(if: $includeAWS) {
      ...AWSConnectForm_setup
      ...AWSConnectInstallActions_install
    }
    azureConnectorSetup(organizationId: $organizationId) @include(if: $includeAzure) {
      ...AzureConnectForm_setup
      ...AzureConnectInstallActions_install
    }
    gcpConnectorSetup(organizationId: $organizationId) @include(if: $includeGCP) {
      ...GCPConnectForm_setup
      ...GCPConnectInstallActions_install
    }
  }
`;

interface ConnectVendorPageProps {
  queryRef: PreloadedQuery<ConnectVendorPageQuery>;
}

export function ConnectVendorPage({ queryRef }: ConnectVendorPageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const { provider: providerSlug = "", method: methodSlug } = useParams();
  const {
    organization,
    connectorProviders,
    awsConnectorSetup,
    azureConnectorSetup,
    gcpConnectorSetup,
  } = usePreloadedQuery<ConnectVendorPageQuery>(connectVendorPageQuery, queryRef);

  const providerId = providerFromSlug(providerSlug);
  const driver = connectorProviders.find(item => item.provider === providerId);
  const method = methodSlug == null ? null : connectMethodFromSlug(methodSlug);

  usePageTitle(driver == null
    ? t("marketplacePage.title")
    : t("marketplacePage.connectTitle", { provider: driver.displayName }));

  if (organization.__typename !== "Organization") {
    throw new NotFoundError(t("listPage.notFound"));
  }
  if (driver == null || (methodSlug != null && method == null)) {
    throw new NotFoundError(t("listPage.notFound"));
  }

  const methods = connectMethods({
    configuredProtocols: driver.configuredProtocols,
    apiKeySupported: driver.apiKeySupported,
    apiKeyManaged: driver.apiKeyManaged,
    clientCredentialsSupported: driver.clientCredentialsSupported,
    workloadIdentitySupported: driver.workloadIdentitySupported,
    installSupported: driver.installSupported,
  });
  if (!organization.canCreateConnector || (method != null && !methods.includes(method))) {
    throw new NotFoundError(t("listPage.notFound"));
  }

  if (method == null) {
    if (methods.length === 1) {
      return (
        <Navigate
          to={connectVendorMethodPath(organizationId, driver.provider, methods[0])}
          replace
        />
      );
    }

    return (
      <ConnectVendorChooser
        organizationId={organizationId}
        providerKey={driver}
      />
    );
  }

  return (
    <ConnectVendorMethod
      organizationId={organizationId}
      providerKey={driver}
      method={method}
      awsSetupKey={awsConnectorSetup}
      azureSetupKey={azureConnectorSetup}
      gcpSetupKey={gcpConnectorSetup}
    />
  );
}
