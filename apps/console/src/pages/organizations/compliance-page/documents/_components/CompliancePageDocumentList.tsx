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

import type { CompliancePageDocumentListFragment$key } from "#/__generated__/core/CompliancePageDocumentListFragment.graphql";

import { CompliancePageDocumentListItem } from "./CompliancePageDocumentListItem";

const fragment = graphql`
  fragment CompliancePageDocumentListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageDocumentListItem_compliancePageFragment
    }
    documents(first: 100 filter: { status: [ACTIVE] }) {
      edges {
        node {
          id
          currentPublishedMajor
          ...CompliancePageDocumentListItem_documentFragment
        }
      }
    }
  }
`;

export function CompliancePageDocumentList(props: { fragmentRef: CompliancePageDocumentListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const { compliancePage, documents } = useFragment<CompliancePageDocumentListFragment$key>(fragment, fragmentRef);
  const publishedDocuments = documents.edges.filter(({ node }) => node.currentPublishedMajor != null);

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {publishedDocuments.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No documents available")}
              </Td>
            </Tr>
          )}
          {publishedDocuments.map(({ node: document }) => (
            <CompliancePageDocumentListItem
              key={document.id}
              compliancePageFragmentRef={compliancePage}
              documentFragmentRef={document}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
};
