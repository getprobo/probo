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

import type { CompliancePageAuditListFragment$key } from "#/__generated__/core/CompliancePageAuditListFragment.graphql";

import { CompliancePageAuditListItem } from "./CompliancePageAuditListItem";

const fragment = graphql`
  fragment CompliancePageAuditListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageAuditListItem_compliancePageFragment
    }
    audits(first: 100) {
      edges {
        node {
          id
          ...CompliancePageAuditListItem_auditFragment
        }
      }
    }
  }
`;

export function CompliancePageAuditList(props: { fragmentRef: CompliancePageAuditListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const { audits, compliancePage } = useFragment<CompliancePageAuditListFragment$key>(fragment, fragmentRef);

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Framework")}</Th>
            <Th>{__("Name")}</Th>
            <Th>{__("Valid Until")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {audits.edges.length === 0 && (
            <Tr>
              <Td colSpan={6} className="text-center text-txt-secondary">
                {__("No audits available")}
              </Td>
            </Tr>
          )}
          {audits.edges.map(({ node: audit }) => (
            <CompliancePageAuditListItem
              key={audit.id}
              auditFragmentRef={audit}
              compliancePageFragmentRef={compliancePage}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
