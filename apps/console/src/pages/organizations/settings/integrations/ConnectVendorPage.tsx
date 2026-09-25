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
  CaretLeftIcon,
  CloudIcon,
  DownloadSimpleIcon,
  GithubLogoIcon,
  KeyIcon,
  PasswordIcon,
  PlugsIcon,
} from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, useParams } from "react-router";

import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import {
  type ConnectMethod,
  connectMethods,
} from "#/pages/organizations/access-reviews/connections/_lib/connectMethods";

import { WorkloadIdentityForm, WorkloadIdentityInstallActions } from "./_components/WorkloadIdentityForm";
import { APIKeyConnectForm } from "./_components/APIKeyConnectForm";
import { ClientCredentialsConnectForm } from "./_components/ClientCredentialsConnectForm";
import { OAuthConnectForm } from "./_components/OAuthConnectForm";
import { StartConnectForm } from "./_components/StartConnectForm";
import {
  connectMethodFromSlug,
  connectVendorMethodPath,
  marketplacePath,
  providerFromSlug,
} from "./_lib/integrationPath";

export const connectVendorPageQuery = graphql`
  query ConnectVendorPageQuery($organizationId: ID!) {
    accessReviewDrivers {
      provider
      displayName
      documentationUrl
      configuredProtocols
      apiKeySupported
      apiKeyManaged
      apiKeyFormat {
        pattern
        example
      }
      apiKeyExtraSettings {
        key
        label
        required
      }
      clientCredentialsSupported
      clientCredentialsTokenUrl
      clientCredentialsExtraSettings {
        key
        label
        required
      }
      workloadIdentitySupported
      installSupported
      oauth2Scopes
    }
    awsConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
      cloudFormationQuickCreateURL
    }
    gcpConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
    }
    azureConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateConnector: permission(action: "core:connector:create")
      }
    }
  }
`;

type Driver = ConnectVendorPageQuery["response"]["accessReviewDrivers"][number];

interface ConnectVendorPageProps {
  queryRef: PreloadedQuery<ConnectVendorPageQuery>;
}

export function ConnectVendorPage({ queryRef }: ConnectVendorPageProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const { provider: providerSlug = "", method: methodSlug } = useParams();
  const { organization, accessReviewDrivers, awsConnectorSetup, gcpConnectorSetup, azureConnectorSetup }
    = usePreloadedQuery<ConnectVendorPageQuery>(connectVendorPageQuery, queryRef);

  const providerId = providerFromSlug(providerSlug);
  const driver = accessReviewDrivers.find(item => item.provider === providerId);
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
      awsConnectorSetup={awsConnectorSetup}
      gcpConnectorSetup={gcpConnectorSetup}
      azureConnectorSetup={azureConnectorSetup}
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
          <Card key={method} variant="soft" size={1} interactive className="relative">
            <Link
              to={connectVendorMethodPath(organizationId, driver.provider, method)}
              underline={false}
              className="absolute inset-0 z-0"
              aria-label={t(`marketplacePage.methods.${method}`)}
            />
            <div className="pointer-events-none flex flex-col items-center gap-3 py-2 text-center">
              <MethodIcon method={method} />
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

function MethodIcon({ method }: { method: ConnectMethod }) {
  const className = "size-8 text-sand-11";
  switch (method) {
    case "API_KEY":
      return <KeyIcon className={className} />;
    case "CLIENT_CREDENTIALS":
      return <PasswordIcon className={className} />;
    case "OAUTH2":
      return <PlugsIcon className={className} />;
    case "GITHUB_APP":
      return <GithubLogoIcon className={className} />;
    case "INSTALL":
      return <DownloadSimpleIcon className={className} />;
    case "WORKLOAD_IDENTITY":
      return <CloudIcon className={className} />;
  }
}

function ConnectVendorMethod({
  organizationId,
  driver,
  method,
  awsConnectorSetup,
  gcpConnectorSetup,
  azureConnectorSetup,
}: {
  organizationId: string;
  driver: Driver;
  method: ConnectMethod;
  awsConnectorSetup: ConnectVendorPageQuery["response"]["awsConnectorSetup"];
  gcpConnectorSetup: ConnectVendorPageQuery["response"]["gcpConnectorSetup"];
  azureConnectorSetup: ConnectVendorPageQuery["response"]["azureConnectorSetup"];
}) {
  const { t } = useTranslation("organizations/settings/integrations");

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
        {method === "WORKLOAD_IDENTITY" && (
          <WorkloadIdentityInstallActions
            provider={driver.provider}
            awsConnectorSetup={awsConnectorSetup}
            gcpConnectorSetup={gcpConnectorSetup}
            azureConnectorSetup={azureConnectorSetup}
          />
        )}
      </div>
      <Card variant="soft" size={2}>
        {method === "WORKLOAD_IDENTITY" && (
          <WorkloadIdentityForm
            organizationId={organizationId}
            provider={driver.provider}
            documentationUrl={driver.documentationUrl}
            awsConnectorSetup={awsConnectorSetup}
            gcpConnectorSetup={gcpConnectorSetup}
            azureConnectorSetup={azureConnectorSetup}
          />
        )}
        {method === "API_KEY" && (
          <APIKeyConnectForm organizationId={organizationId} driver={driver} />
        )}
        {method === "CLIENT_CREDENTIALS" && (
          <ClientCredentialsConnectForm organizationId={organizationId} driver={driver} />
        )}
        {method === "OAUTH2" && (
          <OAuthConnectForm organizationId={organizationId} driver={driver} />
        )}
        {(method === "GITHUB_APP" || method === "INSTALL") && (
          <StartConnectForm
            organizationId={organizationId}
            driver={driver}
            method={method}
          />
        )}
      </Card>
    </div>
  );
}
