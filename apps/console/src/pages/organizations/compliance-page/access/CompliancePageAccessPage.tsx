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
import { Spinner, Table, Tbody, Td, Tr } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAccessPageQuery } from "#/__generated__/core/CompliancePageAccessPageQuery.graphql";

import { CompliancePageAccessList } from "./_components/CompliancePageAccessList";

export const compliancePageAccessPageQuery = graphql`
  query CompliancePageAccessPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: compliancePortal @required(action: THROW) {
          # eslint-disable-next-line relay/unused-fields
          id
          ...CompliancePageAccessListFragment
        }
      }
    }
  }
`;

export function CompliancePageAccessPage(props: { queryRef: PreloadedQuery<CompliancePageAccessPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageAccessPageQuery>(compliancePageAccessPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("External Access")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__(
              "Manage who can access your compliance page",
            )}
          </p>
        </div>
      </div>

      {organization.compliancePage
        ? (
            <CompliancePageAccessList
              fragmentRef={organization.compliancePage}
            />
          )
        : (
            <Table>
              <Tbody>
                <Tr>
                  <Td className="text-center text-txt-tertiary py-8">
                    <Spinner />
                  </Td>
                </Tr>
              </Tbody>
            </Table>
          )}
    </div>
  );
}
