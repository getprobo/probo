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
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ConnectorProviderListItem_provider$key } from "#/__generated__/core/ConnectorProviderListItem_provider.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { connectProviderPath } from "../_lib/integrationPath";
import { integrationSection } from "../variants";

const connectorProviderListItemFragment = graphql`
  fragment ConnectorProviderListItem_provider on ConnectorProviderInfo {
    provider
    displayName
    documentationUrl
  }
`;

interface ConnectorProviderListItemProps {
  providerKey: ConnectorProviderListItem_provider$key;
}

// A catalog row. Connecting still happens under access reviews, which is also
// what creates the source, so this links there rather than opening a dialog of
// its own: a connector made here would have no source and no way to gain one
// until access reviews stops owning that step.
export function ConnectorProviderListItem({
  providerKey,
}: ConnectorProviderListItemProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const provider = useFragment(connectorProviderListItemFragment, providerKey);
  const { item, content, name, trailing } = integrationSection();

  return (
    <li className={item()}>
      <ThirdPartyLogo
        thirdParty={provider.provider}
        className="size-6 shrink-0"
      />
      <div className={content()}>
        <span className={name()}>{provider.displayName}</span>
        {provider.documentationUrl && (
          <a
            href={provider.documentationUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-txt-tertiary underline hover:text-txt-primary"
          >
            {t("listPage.actions.documentation")}
          </a>
        )}
      </div>
      <div className={trailing()}>
        <Button variant="primary" asChild>
          <Link to={connectProviderPath(organizationId)}>
            {t("listPage.actions.connect")}
          </Link>
        </Button>
      </div>
    </li>
  );
}
