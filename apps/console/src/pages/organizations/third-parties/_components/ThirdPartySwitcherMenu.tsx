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

import { PlusIcon } from "@phosphor-icons/react";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { ThirdPartySwitcherMenuQuery } from "#/__generated__/core/ThirdPartySwitcherMenuQuery.graphql";
import { navPanelSwitcher } from "#/pages/organizations/_components/NavPanelSwitcher";

import { ThirdPartySwitcherListItem } from "./ThirdPartySwitcherListItem";

export const thirdPartySwitcherMenuQuery = graphql`
  query ThirdPartySwitcherMenuQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateThirdParty: permission(action: "core:thirdParty:create")
        thirdParties(
          first: 50
          filter: { level: 1 }
          orderBy: { field: CREATED_AT, direction: DESC }
        )
          @connection(key: "ThirdPartySwitcherMenu_thirdParties", filters: [])
          @required(action: THROW) {
          __id
          edges {
            node {
              id
              ...ThirdPartySwitcherListItem_thirdParty
            }
          }
        }
      }
    }
  }
`;

export interface ThirdPartySwitcherMenuProps {
  queryRef: PreloadedQuery<ThirdPartySwitcherMenuQuery>;
  onCreate: (connectionId: string) => void;
}

export function ThirdPartySwitcherMenu({ queryRef, onCreate }: ThirdPartySwitcherMenuProps) {
  const { t } = useTranslation();
  const slots = navPanelSwitcher();

  const { organization } = usePreloadedQuery<ThirdPartySwitcherMenuQuery>(
    thirdPartySwitcherMenuQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const thirdParties = organization.thirdParties.edges.map(edge => edge.node);

  return (
    <>
      <div className={slots.list()}>
        {thirdParties.length === 0
          ? (
              <Text size={2} color="faint" className={slots.empty()}>
                {t("nav.thirdPartySwitcher.empty")}
              </Text>
            )
          : thirdParties.map(thirdParty => (
              <ThirdPartySwitcherListItem key={thirdParty.id} thirdPartyKey={thirdParty} />
            ))}
      </div>
      {organization.canCreateThirdParty && (
        <>
          <DropdownSeparator />
          <DropdownItem
            iconStart={<PlusIcon />}
            onClick={() => {
              onCreate(organization.thirdParties.__id);
            }}
          >
            {t("nav.thirdPartySwitcher.create")}
          </DropdownItem>
        </>
      )}
    </>
  );
}
