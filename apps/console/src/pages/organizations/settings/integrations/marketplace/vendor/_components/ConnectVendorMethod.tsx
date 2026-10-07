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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type {
  ConnectVendorMethod_provider$data,
  ConnectVendorMethod_provider$key,
} from "#/__generated__/core/ConnectVendorMethod_provider.graphql";
import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import type { ConnectMethod } from "../../../_lib/connectMethods";
import { marketplacePath } from "../../../_lib/integrationPath";

import { APIKeyConnectForm } from "./APIKeyConnectForm";
import { AWSConnectForm } from "./AWSConnectForm";
import { AWSConnectInstallActions } from "./AWSConnectInstallActions";
import { AzureConnectForm } from "./AzureConnectForm";
import { AzureConnectInstallActions } from "./AzureConnectInstallActions";
import { ClientCredentialsConnectForm } from "./ClientCredentialsConnectForm";
import { DatadogConnectForm } from "./DatadogConnectForm";
import { GCPConnectForm } from "./GCPConnectForm";
import { GCPConnectInstallActions } from "./GCPConnectInstallActions";
import { OAuthConnectForm } from "./OAuthConnectForm";
import { StartConnectForm } from "./StartConnectForm";
import { ZendeskConnectForm } from "./ZendeskConnectForm";

const connectVendorMethodFragment = graphql`
  fragment ConnectVendorMethod_provider on ConnectorProviderInfo {
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

interface ConnectVendorMethodProps {
  providerKey: ConnectVendorMethod_provider$key;
  method: ConnectMethod;
  awsSetupKey: ConnectVendorPageQuery["response"]["awsConnectorSetup"];
  azureSetupKey: ConnectVendorPageQuery["response"]["azureConnectorSetup"];
  gcpSetupKey: ConnectVendorPageQuery["response"]["gcpConnectorSetup"];
}

export function ConnectVendorMethod({
  providerKey,
  method,
  awsSetupKey,
  azureSetupKey,
  gcpSetupKey,
}: ConnectVendorMethodProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const driver = useFragment(connectVendorMethodFragment, providerKey);
  const cloud = workloadIdentityCloud({
    providerKey: driver,
    provider: driver.provider,
    method,
    awsSetupKey,
    azureSetupKey,
    gcpSetupKey,
    notFoundMessage: t("listPage.notFound"),
  });
  const oauth = oauthConnectForm({
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
          <APIKeyConnectForm providerKey={driver} />
        )}
        {method === "CLIENT_CREDENTIALS" && (
          <ClientCredentialsConnectForm providerKey={driver} />
        )}
        {oauth}
        {(method === "GITHUB_APP" || method === "INSTALL") && (
          <StartConnectForm
            providerKey={driver}
            method={method}
          />
        )}
      </Card>
    </div>
  );
}

function oauthConnectForm({
  providerKey,
  provider,
  method,
}: {
  providerKey: ConnectVendorMethod_provider$data;
  provider: ConnectVendorMethod_provider$data["provider"];
  method: ConnectMethod;
}): ReactNode | null {
  if (method !== "OAUTH2") {
    return null;
  }
  if (provider === "DATADOG") {
    return <DatadogConnectForm providerKey={providerKey} />;
  }
  if (provider === "ZENDESK") {
    return <ZendeskConnectForm providerKey={providerKey} />;
  }

  return <OAuthConnectForm providerKey={providerKey} />;
}

function workloadIdentityCloud({
  providerKey,
  provider,
  method,
  awsSetupKey,
  azureSetupKey,
  gcpSetupKey,
  notFoundMessage,
}: {
  providerKey: ConnectVendorMethod_provider$data;
  provider: ConnectVendorMethod_provider$data["provider"];
  method: ConnectMethod;
  awsSetupKey: ConnectVendorPageQuery["response"]["awsConnectorSetup"];
  azureSetupKey: ConnectVendorPageQuery["response"]["azureConnectorSetup"];
  gcpSetupKey: ConnectVendorPageQuery["response"]["gcpConnectorSetup"];
  notFoundMessage: string;
}): { actions: ReactNode; form: ReactNode } | null {
  if (method !== "WORKLOAD_IDENTITY") {
    return null;
  }
  if (provider === "AWS") {
    if (awsSetupKey == null) {
      throw new NotFoundError(notFoundMessage);
    }

    return {
      actions: <AWSConnectInstallActions installKey={awsSetupKey} />,
      form: (
        <AWSConnectForm
          providerKey={providerKey}
          setupKey={awsSetupKey}
        />
      ),
    };
  }
  if (provider === "GCP") {
    if (gcpSetupKey == null) {
      throw new NotFoundError(notFoundMessage);
    }

    return {
      actions: <GCPConnectInstallActions installKey={gcpSetupKey} />,
      form: (
        <GCPConnectForm
          providerKey={providerKey}
          setupKey={gcpSetupKey}
        />
      ),
    };
  }
  if (provider === "AZURE") {
    if (azureSetupKey == null) {
      throw new NotFoundError(notFoundMessage);
    }

    return {
      actions: <AzureConnectInstallActions installKey={azureSetupKey} />,
      form: (
        <AzureConnectForm
          providerKey={providerKey}
          setupKey={azureSetupKey}
        />
      ),
    };
  }

  return null;
}
