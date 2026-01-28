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
                {__("No vendors available")}
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
