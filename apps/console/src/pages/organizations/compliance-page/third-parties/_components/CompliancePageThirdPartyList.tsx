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

import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageThirdPartyListFragment$key } from "#/__generated__/core/CompliancePageThirdPartyListFragment.graphql";

import { CompliancePageThirdPartyListItem } from "./CompliancePageThirdPartyListItem";

const fragment = graphql`
  fragment CompliancePageThirdPartyListFragment on Organization {
    thirdParties(first: 100) {
      edges {
        node {
          id
          ...CompliancePageThirdPartyListItem_thirdPartyFragment
        }
      }
    }
  }
`;

export function CompliancePageThirdPartyList(props: { fragmentRef: CompliancePageThirdPartyListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const { thirdParties } = useFragment<CompliancePageThirdPartyListFragment$key>(fragment, fragmentRef);

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Category")}</Th>
            <Th>{__("Visibility")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {thirdParties.edges.length === 0 && (
            <Tr>
              <Td colSpan={4} className="text-center text-txt-secondary">
                {__("No subprocessors available")}
              </Td>
            </Tr>
          )}
          {thirdParties.edges.map(({ node: thirdParty }) => (
            <CompliancePageThirdPartyListItem
              key={thirdParty.id}
              thirdPartyFragmentRef={thirdParty}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
