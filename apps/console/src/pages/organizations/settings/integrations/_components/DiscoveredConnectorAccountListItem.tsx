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

import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DiscoveredConnectorAccountListItem_account$key } from "#/__generated__/core/DiscoveredConnectorAccountListItem_account.graphql";
import type { DiscoveredConnectorAccountListItem_connector$key } from "#/__generated__/core/DiscoveredConnectorAccountListItem_connector.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import { connectorCard } from "../variants";

const discoveredConnectorAccountListItemFragment = graphql`
  fragment DiscoveredConnectorAccountListItem_account on DiscoveredConnectorAccount {
    externalAccountId
    name
  }
`;

const discoveredConnectorAccountListItemConnectorFragment = graphql`
  fragment DiscoveredConnectorAccountListItem_connector on Connector {
    provider
  }
`;

interface DiscoveredConnectorAccountListItemProps {
  accountKey: DiscoveredConnectorAccountListItem_account$key;
  connectorKey: DiscoveredConnectorAccountListItem_connector$key;
  selected: boolean;
  selectable: boolean;
  onSelectedChange: (checked: boolean) => void;
}

export function DiscoveredConnectorAccountListItem({
  accountKey,
  connectorKey,
  selected,
  selectable,
  onSelectedChange,
}: DiscoveredConnectorAccountListItemProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const account = useFragment(discoveredConnectorAccountListItemFragment, accountKey);
  const connector = useFragment(discoveredConnectorAccountListItemConnectorFragment, connectorKey);
  const { identity, name, title } = connectorCard();

  return (
    <TonedCard
      tone="sand"
      size={2}
      icon={(
        <ThirdPartyLogo thirdParty={connector.provider} />
      )}
      lead={(
        <div className={identity()}>
          <div className={name()}>
            <Heading level={2} size={3} weight="medium" highContrast className={title()}>
              {account.name}
            </Heading>
            <Text size={1} color="faint" className="shrink-0 font-mono">
              #
              {account.externalAccountId}
            </Text>
          </div>
          <Badge variant="soft" color="neutral" size={1}>
            {t("detailsPage.accounts.pending")}
          </Badge>
        </div>
      )}
      control={selectable
        ? (
            <Checkbox
              checked={selected}
              aria-label={account.name}
              onCheckedChange={onSelectedChange}
            />
          )
        : undefined}
    />
  );
}
