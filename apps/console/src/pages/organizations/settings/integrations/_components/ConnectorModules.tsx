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

import { GearIcon, KeyIcon, type Icon } from "@phosphor-icons/react";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { ConnectorModulesQuery } from "#/__generated__/core/ConnectorModulesQuery.graphql";

import { connectorCard } from "../variants";

const connectorModulesQuery = graphql`
  query ConnectorModulesQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        connectors {
          id
          modules
        }
      }
    }
  }
`;

function moduleIcon(module: string): Icon | null {
  switch (module) {
    case "ACCESS_REVIEW":
      return KeyIcon;
    case "SCIM":
      return GearIcon;
    default:
      return null;
  }
}

interface ConnectorModulesProps {
  connectorId: string;
  organizationId: string;
}

export function ConnectorModules({
  connectorId,
  organizationId,
}: ConnectorModulesProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const data = useLazyLoadQuery<ConnectorModulesQuery>(
    connectorModulesQuery,
    { organizationId },
    { fetchPolicy: "store-and-network" },
  );
  const { usedBy, usedByIcons } = connectorCard();

  if (data.node?.__typename !== "Organization") {
    return null;
  }

  const modules = data.node.connectors
    .find(connector => connector.id === connectorId)
    ?.modules
    .flatMap((module) => {
      const IconComponent = moduleIcon(module);
      if (IconComponent == null) {
        return [];
      }

      const label = t(`listPage.modules.${module}`);

      return [(
        <Tooltip key={module}>
          <TooltipTrigger
            aria-label={label}
            className="text-sand-11"
          >
            <IconComponent className="size-4" />
          </TooltipTrigger>
          <TooltipPopup>{label}</TooltipPopup>
        </Tooltip>
      )];
    }) ?? [];

  if (modules.length === 0) {
    return (
      <Text size={1} color="faint">
        {t("listPage.notUsedYet")}
      </Text>
    );
  }

  return (
    <div className={usedBy()}>
      <Text size={1} color="faint">
        {t("listPage.usedBy")}
      </Text>
      <div className={usedByIcons()}>{modules}</div>
    </div>
  );
}

export function ConnectorModulesSkeleton() {
  const { usedBy } = connectorCard();

  return (
    <div className={usedBy()} aria-hidden>
      <TextSkeleton size={1} className="w-24" />
    </div>
  );
}
