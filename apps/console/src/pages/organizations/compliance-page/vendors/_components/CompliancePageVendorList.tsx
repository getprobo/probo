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
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageVendorListFragment$key } from "#/__generated__/core/CompliancePageVendorListFragment.graphql";

import { CompliancePageVendorListItem } from "./CompliancePageVendorListItem";

const fragment = graphql`
  fragment CompliancePageVendorListFragment on Organization {
    vendors(first: 100) {
      edges {
        node {
          id
          ...CompliancePageVendorListItem_vendorFragment
        }
      }
    }
  }
`;

export function CompliancePageVendorList(props: { fragmentRef: CompliancePageVendorListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const { vendors } = useFragment<CompliancePageVendorListFragment$key>(fragment, fragmentRef);

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
          {vendors.edges.length === 0 && (
            <Tr>
              <Td colSpan={4} className="text-center text-txt-secondary">
                {__("No subprocessors available")}
              </Td>
            </Tr>
          )}
          {vendors.edges.map(({ node: vendor }) => (
            <CompliancePageVendorListItem
              key={vendor.id}
              vendorFragmentRef={vendor}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
