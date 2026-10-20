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
import { CardLink } from "@probo/ui/src/v2/Card/CardLink";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectVendorChooser_provider$key } from "#/__generated__/core/ConnectVendorChooser_provider.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ConnectorMethodIcon } from "../../../_components/ConnectorMethodIcon";
import { connectMethods, connectVendorMethodPath } from "../../../_lib/connectMethods";
import { marketplacePath } from "../../../_lib/integrationPath";

const connectVendorChooserFragment = graphql`
  fragment ConnectVendorChooser_provider on ConnectorProviderInfo {
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

interface ConnectVendorChooserProps {
  providerKey: ConnectVendorChooser_provider$key;
}

export function ConnectVendorChooser({
  providerKey,
}: ConnectVendorChooserProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const driver = useFragment(connectVendorChooserFragment, providerKey);
  const methods = connectMethods({
    configuredProtocols: driver.configuredProtocols,
    apiKeySupported: driver.apiKeySupported,
    apiKeyManaged: driver.apiKeyManaged,
    clientCredentialsSupported: driver.clientCredentialsSupported,
    workloadIdentitySupported: driver.workloadIdentitySupported,
    installSupported: driver.installSupported,
  });

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
