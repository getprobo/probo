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

import type { ConnectorListItem_connector$key } from "#/__generated__/core/ConnectorListItem_connector.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { ConnectorDocumentationLink } from "#/pages/organizations/access-reviews/dialogs/_components/ConnectorDocumentationLink";

import { connectorDetailsPath } from "../_lib/integrationPath";
import { integrationSection } from "../variants";

// connectionStatus and accounts stay off this list. connectionStatus probes
// the vendor on every read, and accounts authorizes core:connector:get, which
// a role that can only list connectors does not hold.
const connectorListItemFragment = graphql`
  fragment ConnectorListItem_connector on Connector {
    id
    provider
    displayName
    documentationUrl
    canGet: permission(action: "core:connector:get")
  }
`;

interface ConnectorListItemProps {
  connectorKey: ConnectorListItem_connector$key;
}

export function ConnectorListItem({ connectorKey }: ConnectorListItemProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const connector = useFragment(connectorListItemFragment, connectorKey);
  const { item, content, trailing } = integrationSection();

  return (
    <li className={item()}>
      <ThirdPartyLogo
        thirdParty={connector.provider}
        className="size-6 shrink-0"
      />
      <div className={content()}>
        <span className="text-sm font-medium text-txt-primary">
          {connector.displayName}
        </span>
        <ConnectorDocumentationLink url={connector.documentationUrl} />
      </div>
      {connector.canGet && (
        <div className={trailing()}>
          <Button variant="primary" asChild>
            <Link to={connectorDetailsPath(organizationId, connector.id)}>
              {t("listPage.actions.manage")}
            </Link>
          </Button>
        </div>
      )}
    </li>
  );
}
