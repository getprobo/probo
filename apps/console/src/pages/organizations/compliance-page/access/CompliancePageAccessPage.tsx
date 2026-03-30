// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
        compliancePage: trustCenter @required(action: THROW) {
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
