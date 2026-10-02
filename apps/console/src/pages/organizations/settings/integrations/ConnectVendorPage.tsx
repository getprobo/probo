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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { Navigate, useParams } from "react-router";

import type { ConnectVendorPageChooser_provider$key } from "#/__generated__/core/ConnectVendorPageChooser_provider.graphql";
import type {
  ConnectVendorPageMethod_provider$data,
  ConnectVendorPageMethod_provider$key,
} from "#/__generated__/core/ConnectVendorPageMethod_provider.graphql";
import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { APIKeyConnectForm } from "./_components/APIKeyConnectForm";
import { AWSConnectForm, AWSConnectInstallActions } from "./_components/AWSConnectForm";
import { AzureConnectForm, AzureConnectInstallActions } from "./_components/AzureConnectForm";
import { ClientCredentialsConnectForm } from "./_components/ClientCredentialsConnectForm";
import { ConnectorMethodIcon } from "./_components/ConnectorMethodIcon";
import { DatadogConnectForm } from "./_components/DatadogConnectForm";
import { GCPConnectForm, GCPConnectInstallActions } from "./_components/GCPConnectForm";
import { OAuthConnectForm } from "./_components/OAuthConnectForm";
import { StartConnectForm } from "./_components/StartConnectForm";
import { ZendeskConnectForm } from "./_components/ZendeskConnectForm";
import {
  type ConnectMethod,
  connectMethods,
} from "./_lib/connectMethods";
import {
  connectMethodFromSlug,
  connectVendorMethodPath,
  marketplacePath,
  providerFromSlug,
} from "./_lib/integrationPath";

const connectVendorChooserFragment = graphql`
  fragment ConnectVendorPageChooser_provider on ConnectorProviderInfo {
    provider
    displayName
    configuredProtocols
    apiKeySupported
    apiKeyManaged
    clientCredentialsSupported
    workloadIdentitySupported
    installSupported
  }
`;

const connectVendorMethodFragment = graphql`
  fragment ConnectVendorPageMethod_provider on ConnectorProviderInfo {
    provider
    displayName
    ...APIKeyConnectForm_provider
    ...ClientCredentialsConnectForm_provider
    ...OAuthConnectForm_provider
    ...StartConnectForm_provider
    ...AWSConnectForm_provider
    ...GCPConnectForm_provider
    ...AzureConnectForm_provider
  }
`;

export const connectVendorPageQuery = graphql`
  query ConnectVendorPageQuery($organizationId: ID!) {
    connectorProviders {
      provider
      displayName
      configuredProtocols
      apiKeySupported
      apiKeyManaged
      clientCredentialsSupported
      workloadIdentitySupported
      installSupported
      ...ConnectVendorPageChooser_provider
      ...ConnectVendorPageMethod_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
      }
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
  const { organization, connectorProviders }
    = usePreloadedQuery<ConnectVendorPageQuery>(connectVendorPageQuery, queryRef);

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

  const methods = providerConnectMethods(driver);
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
    />
  );
}

function ConnectVendorChooser({
  organizationId,
  providerKey,
}: {
  organizationId: string;
  providerKey: ConnectVendorPageChooser_provider$key;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const driver = useFragment(connectVendorChooserFragment, providerKey);
  const methods = providerConnectMethods(driver);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <Link
          to={marketplacePath(organizationId)}
          size={2}
          color="neutral"
          underline={false}
          iconStart={<CaretLeftIcon />}
          className="self-start"
        >
          {t("marketplacePage.title")}
        </Link>
        <Heading level={1} size={6} weight="medium" highContrast>
          {t("marketplacePage.connectTitle", { provider: driver.displayName })}
        </Heading>
        <Text size={2} color="faint">
          {t("marketplacePage.choose", { provider: driver.displayName })}
        </Text>
      </div>
      <div className="grid grid-cols-4 gap-3 max-xl:grid-cols-3 max-lg:grid-cols-2 max-sm:grid-cols-1">
        {methods.map(method => (
          <Card key={method} variant="soft" size={1} interactive className="relative">
            <Link
              to={connectVendorMethodPath(organizationId, driver.provider, method)}
              underline={false}
              className="absolute inset-0 z-0"
              aria-label={t(`marketplacePage.methods.${method}`)}
            />
            <div className="pointer-events-none flex flex-col items-center gap-3 py-2 text-center">
              <ConnectorMethodIcon method={method} className="size-8 text-sand-11" />
              <Heading level={2} size={2} weight="medium" highContrast>
                {t(`marketplacePage.methods.${method}`)}
              </Heading>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

function ConnectVendorMethod({
  organizationId,
  providerKey,
  method,
}: {
  organizationId: string;
  providerKey: ConnectVendorPageMethod_provider$key;
  method: ConnectMethod;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const driver = useFragment(connectVendorMethodFragment, providerKey);
  const cloud = workloadIdentityCloud({
    organizationId,
    providerKey: driver,
    provider: driver.provider,
    method,
  });
  const oauth = oauthConnectForm({
    organizationId,
    providerKey: driver,
    provider: driver.provider,
    method,
  });

  return (
    <div className="flex flex-col gap-6">
      <Link
        to={marketplacePath(organizationId)}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className="self-start"
      >
        {t("marketplacePage.title")}
      </Link>
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 flex-col gap-2">
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("marketplacePage.connectTitle", { provider: driver.displayName })}
          </Heading>
          <Text size={2} color="faint">
            {t(`marketplacePage.methodDescriptions.${method}`)}
          </Text>
        </div>
        {cloud?.actions}
      </div>
      <Card variant="soft" size={2}>
        {cloud?.form}
        {method === "API_KEY" && (
          <APIKeyConnectForm organizationId={organizationId} providerKey={driver} />
        )}
        {method === "CLIENT_CREDENTIALS" && (
          <ClientCredentialsConnectForm organizationId={organizationId} providerKey={driver} />
        )}
        {oauth}
        {(method === "GITHUB_APP" || method === "INSTALL") && (
          <StartConnectForm
            organizationId={organizationId}
            providerKey={driver}
            method={method}
          />
        )}
      </Card>
    </div>
  );
}

function providerConnectMethods(provider: {
  configuredProtocols: ConnectVendorPageQuery["response"]["connectorProviders"][number]["configuredProtocols"];
  apiKeySupported: boolean;
  apiKeyManaged: boolean;
  clientCredentialsSupported: boolean;
  workloadIdentitySupported: boolean;
  installSupported: boolean;
}) {
  return connectMethods({
    configuredProtocols: provider.configuredProtocols,
    apiKeySupported: provider.apiKeySupported,
    apiKeyManaged: provider.apiKeyManaged,
    clientCredentialsSupported: provider.clientCredentialsSupported,
    workloadIdentitySupported: provider.workloadIdentitySupported,
    installSupported: provider.installSupported,
  });
}

function oauthConnectForm({
  organizationId,
  providerKey,
  provider,
  method,
}: {
  organizationId: string;
  providerKey: ConnectVendorPageMethod_provider$data;
  provider: ConnectVendorPageMethod_provider$data["provider"];
  method: ConnectMethod;
}): ReactNode | null {
  if (method !== "OAUTH2") {
    return null;
  }
  if (provider === "DATADOG") {
    return (
      <DatadogConnectForm
        organizationId={organizationId}
        providerKey={providerKey}
      />
    );
  }
  if (provider === "ZENDESK") {
    return (
      <ZendeskConnectForm
        organizationId={organizationId}
        providerKey={providerKey}
      />
    );
  }

  return (
    <OAuthConnectForm
      organizationId={organizationId}
      providerKey={providerKey}
    />
  );
}

function workloadIdentityCloud({
  organizationId,
  providerKey,
  provider,
  method,
}: {
  organizationId: string;
  providerKey: ConnectVendorPageMethod_provider$data;
  provider: ConnectVendorPageMethod_provider$data["provider"];
  method: ConnectMethod;
}): { actions: ReactNode; form: ReactNode } | null {
  if (method !== "WORKLOAD_IDENTITY") {
    return null;
  }
  if (provider === "AWS") {
    return {
      actions: <AWSConnectInstallActions organizationId={organizationId} />,
      form: (
        <AWSConnectForm
          organizationId={organizationId}
          providerKey={providerKey}
        />
      ),
    };
  }
  if (provider === "GCP") {
    return {
      actions: <GCPConnectInstallActions organizationId={organizationId} />,
      form: (
        <GCPConnectForm
          organizationId={organizationId}
          providerKey={providerKey}
        />
      ),
    };
  }
  if (provider === "AZURE") {
    return {
      actions: <AzureConnectInstallActions organizationId={organizationId} />,
      form: (
        <AzureConnectForm
          organizationId={organizationId}
          providerKey={providerKey}
        />
      ),
    };
  }

  return null;
}
