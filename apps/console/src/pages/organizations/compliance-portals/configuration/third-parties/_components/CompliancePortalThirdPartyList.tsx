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

import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalThirdPartyList_compliancePortal$key } from "#/__generated__/core/CompliancePortalThirdPartyList_compliancePortal.graphql";
import type { CompliancePortalThirdPartyList_organization$key } from "#/__generated__/core/CompliancePortalThirdPartyList_organization.graphql";
import type {
  CompliancePortalThirdPartyListPaginationQuery,
  ThirdPartyOrderField,
} from "#/__generated__/core/CompliancePortalThirdPartyListPaginationQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CompliancePortalThirdPartyListItem } from "./CompliancePortalThirdPartyListItem";

const compliancePortalFragment = graphql`
  fragment CompliancePortalThirdPartyList_compliancePortal on CompliancePortal {
    id
    ...CompliancePortalThirdPartyListItem_compliancePortal
  }
`;

const organizationFragment = graphql`
  fragment CompliancePortalThirdPartyList_organization on Organization
  @refetchable(queryName: "CompliancePortalThirdPartyListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    order: {
      type: "ThirdPartyOrder"
      defaultValue: { field: NAME, direction: ASC }
    }
    compliancePortalId: { type: "ID!" }
  ) {
    thirdParties(
      first: $first
      after: $after
      before: $before
      last: $last
      orderBy: $order
      filter: { level: 1 }
    )
      @connection(
        key: "CompliancePortalThirdPartyList_thirdParties"
        filters: ["orderBy", "filter"]
      ) {
      edges {
        node {
          id
          ...CompliancePortalThirdPartyListItem_thirdParty
            @arguments(compliancePortalId: $compliancePortalId)
        }
      }
    }
  }
`;

export function CompliancePortalThirdPartyList(props: {
  organizationKey: CompliancePortalThirdPartyList_organization$key;
  compliancePortalKey: CompliancePortalThirdPartyList_compliancePortal$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");
  const compliancePortal = useFragment(
    compliancePortalFragment,
    props.compliancePortalKey,
  );
  const pagination = usePaginationFragment<
    CompliancePortalThirdPartyListPaginationQuery,
    CompliancePortalThirdPartyList_organization$key
  >(organizationFragment, props.organizationKey);

  const thirdParties = pagination.data.thirdParties.edges.map(({ node }) => node);
  const refetch: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      compliancePortalId: compliancePortal.id,
      order: {
        direction: order.direction,
        field: order.field as ThirdPartyOrderField,
      },
    });
  };

  return (
    <div className="space-y-2.5">
      <SortableTable
        {...pagination}
        refetch={refetch}
        initialOrder={{ field: "NAME", direction: "ASC" }}
      >
        <Thead>
          <Tr>
            <Th className="w-24">{t("thirdPartyList.columns.displayed")}</Th>
            <SortableTh field="NAME">
              {t("thirdPartyList.columns.name")}
            </SortableTh>
            <Th>{t("thirdPartyList.columns.category")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {thirdParties.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {t("thirdPartyList.empty")}
              </Td>
            </Tr>
          )}
          {thirdParties.map(thirdParty => (
            <CompliancePortalThirdPartyListItem
              key={thirdParty.id}
              thirdPartyKey={thirdParty}
              compliancePortalKey={compliancePortal}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
