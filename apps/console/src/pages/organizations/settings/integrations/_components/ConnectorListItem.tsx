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

import { connectorDetailsPath } from "../_lib/integrationPath";
import { integrationSection } from "../variants";

// connectionStatus is deliberately absent: it probes the vendor on every read,
// so selecting it here would cost one outbound call per connected integration
// every time the list renders. The details page asks for one connector.
const connectorListItemFragment = graphql`
  fragment ConnectorListItem_connector on Connector {
    id
    provider
    displayName
    accounts(first: 0) {
      totalCount
    }
  }
`;

interface ConnectorListItemProps {
  connectorKey: ConnectorListItem_connector$key;
}

export function ConnectorListItem({ connectorKey }: ConnectorListItemProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const connector = useFragment(connectorListItemFragment, connectorKey);
  const { item, content, name, description, trailing } = integrationSection();

  return (
    <li className={item()}>
      <ThirdPartyLogo
        thirdParty={connector.provider}
        className="size-6 shrink-0"
      />
      <div className={content()}>
        <span className={name()}>{connector.displayName}</span>
        <span className={description()}>
          {t("listPage.accountCount", {
            count: connector.accounts.totalCount,
          })}
        </span>
      </div>
      <div className={trailing()}>
        <Button variant="secondary" asChild>
          <Link to={connectorDetailsPath(organizationId, connector.id)}>
            {t("listPage.actions.manage")}
          </Link>
        </Button>
      </div>
    </li>
  );
}
