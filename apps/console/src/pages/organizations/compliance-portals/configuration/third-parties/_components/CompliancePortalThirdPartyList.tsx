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

import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalThirdPartyListFragment$key } from "#/__generated__/core/CompliancePortalThirdPartyListFragment.graphql";

import { CompliancePortalThirdPartyListItem } from "./CompliancePortalThirdPartyListItem";
import { NewCompliancePortalThirdPartyDialog } from "./NewCompliancePortalThirdPartyDialog";

const fragment = graphql`
  fragment CompliancePortalThirdPartyListFragment on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
    thirdParties(first: 100)
      @connection(key: "CompliancePortalThirdPartyList_thirdParties") {
      __id
      edges {
        node {
          id
          ...CompliancePortalThirdPartyListItem_catalogThirdPartyFragment
        }
      }
    }
  }
`;

export function CompliancePortalThirdPartyList(props: {
  fragmentRef: CompliancePortalThirdPartyListFragment$key;
}) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-portals");

  const compliancePortal = useFragment<CompliancePortalThirdPartyListFragment$key>(fragment, fragmentRef);

  return (
    <div className="space-y-[10px]">
      <div className="flex justify-end">
        <NewCompliancePortalThirdPartyDialog
          compliancePortalId={compliancePortal.id}
          connectionId={compliancePortal.thirdParties.__id}
          disabled={!compliancePortal.canUpdate}
        />
      </div>
      <Table>
        <Thead>
          <Tr>
            <Th>{t("thirdPartyList.columns.name")}</Th>
            <Th>{t("thirdPartyList.columns.category")}</Th>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {compliancePortal.thirdParties.edges.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {t("thirdPartyList.empty")}
              </Td>
            </Tr>
          )}
          {compliancePortal.thirdParties.edges.map(({ node: catalogThirdParty }) => (
            <CompliancePortalThirdPartyListItem
              key={catalogThirdParty.id}
              catalogThirdPartyFragmentRef={catalogThirdParty}
              thirdPartiesConnectionId={compliancePortal.thirdParties.__id}
              canUpdatePortal={compliancePortal.canUpdate}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
