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
import { CardLink } from "@probo/ui/src/v2/Card/CardLink";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, useParams } from "react-router";

import type { AWSConnectFormQuery } from "#/__generated__/core/AWSConnectFormQuery.graphql";
import type { AzureConnectFormQuery } from "#/__generated__/core/AzureConnectFormQuery.graphql";
import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import type { GCPConnectFormQuery } from "#/__generated__/core/GCPConnectFormQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { APIKeyConnectForm } from "./_components/APIKeyConnectForm";
import { AWSConnectForm, AWSConnectInstallActions } from "./_components/AWSConnectForm";
import { AzureConnectForm, AzureConnectInstallActions } from "./_components/AzureConnectForm";
import { ClientCredentialsConnectForm } from "./_components/ClientCredentialsConnectForm";
import { ConnectorMethodIcon } from "./_components/ConnectorMethodIcon";
import { GCPConnectForm, GCPConnectInstallActions } from "./_components/GCPConnectForm";
import { OAuthConnectForm } from "./_components/OAuthConnectForm";
import { StartConnectForm } from "./_components/StartConnectForm";
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

export const connectVendorPageQuery = graphql`
  query ConnectVendorPageQuery($organizationId: ID!) {
    connectorProviders {
      provider
      displayName
      documentationUrl
      configuredProtocols
      apiKeySupported
      apiKeyManaged
      clientCredentialsSupported
      workloadIdentitySupported
      installSupported
      ...APIKeyConnectForm_provider
      ...ClientCredentialsConnectForm_provider
      ...OAuthConnectForm_provider
      ...StartConnectForm_provider
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
      }
    }
  }
`;

type Driver = ConnectVendorPageQuery["response"]["connectorProviders"][number];

interface ConnectVendorPageProps {
  queryRef: PreloadedQuery<ConnectVendorPageQuery>;
  awsQueryRef: PreloadedQuery<AWSConnectFormQuery> | null;
  gcpQueryRef: PreloadedQuery<GCPConnectFormQuery> | null;
  azureQueryRef: PreloadedQuery<AzureConnectFormQuery> | null;
}

export function ConnectVendorPage({
  queryRef,
  awsQueryRef,
  gcpQueryRef,
  azureQueryRef,
}: ConnectVendorPageProps) {
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

  const methods = connectMethods(driver);
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
        driver={driver}
        methods={methods}
      />
    );
  }

  return (
    <ConnectVendorMethod
      organizationId={organizationId}
      driver={driver}
      method={method}
      awsQueryRef={awsQueryRef}
      gcpQueryRef={gcpQueryRef}
      azureQueryRef={azureQueryRef}
    />
  );
}

function ConnectVendorChooser({
  organizationId,
  driver,
  methods,
}: {
  organizationId: string;
  driver: Driver;
  methods: ReadonlyArray<ConnectMethod>;
}) {
  const { t } = useTranslation("organizations/settings/integrations");

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
          <Card key={method} variant="soft" size={1} interactive>
            <CardLink
              to={connectVendorMethodPath(organizationId, driver.provider, method)}
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
  driver,
  method,
  awsQueryRef,
  gcpQueryRef,
  azureQueryRef,
}: {
  organizationId: string;
  driver: Driver;
  method: ConnectMethod;
  awsQueryRef: PreloadedQuery<AWSConnectFormQuery> | null;
  gcpQueryRef: PreloadedQuery<GCPConnectFormQuery> | null;
  azureQueryRef: PreloadedQuery<AzureConnectFormQuery> | null;
}) {
  const { t } = useTranslation("organizations/settings/integrations");
  const cloud = workloadIdentityCloud({
    organizationId,
    documentationUrl: driver.documentationUrl,
    provider: driver.provider,
    method,
    awsQueryRef,
    gcpQueryRef,
    azureQueryRef,
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
        {method === "OAUTH2" && (
          <OAuthConnectForm organizationId={organizationId} providerKey={driver} />
        )}
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

function workloadIdentityCloud({
  organizationId,
  documentationUrl,
  provider,
  method,
  awsQueryRef,
  gcpQueryRef,
  azureQueryRef,
}: {
  organizationId: string;
  documentationUrl: string | null | undefined;
  provider: Driver["provider"];
  method: ConnectMethod;
  awsQueryRef: PreloadedQuery<AWSConnectFormQuery> | null;
  gcpQueryRef: PreloadedQuery<GCPConnectFormQuery> | null;
  azureQueryRef: PreloadedQuery<AzureConnectFormQuery> | null;
}): { actions: ReactNode; form: ReactNode } | null {
  if (method !== "WORKLOAD_IDENTITY") {
    return null;
  }
  if (provider === "AWS" && awsQueryRef != null) {
    return {
      actions: <AWSConnectInstallActions queryRef={awsQueryRef} />,
      form: (
        <AWSConnectForm
          organizationId={organizationId}
          documentationUrl={documentationUrl}
          queryRef={awsQueryRef}
        />
      ),
    };
  }
  if (provider === "GCP" && gcpQueryRef != null) {
    return {
      actions: <GCPConnectInstallActions queryRef={gcpQueryRef} />,
      form: (
        <GCPConnectForm
          organizationId={organizationId}
          documentationUrl={documentationUrl}
          queryRef={gcpQueryRef}
        />
      ),
    };
  }
  if (provider === "AZURE" && azureQueryRef != null) {
    return {
      actions: <AzureConnectInstallActions queryRef={azureQueryRef} />,
      form: (
        <AzureConnectForm
          organizationId={organizationId}
          documentationUrl={documentationUrl}
          queryRef={azureQueryRef}
        />
      ),
    };
  }

  return null;
}
