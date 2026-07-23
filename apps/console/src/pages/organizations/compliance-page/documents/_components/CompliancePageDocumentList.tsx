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

import type { CompliancePageDocumentListFragment$key } from "#/__generated__/core/CompliancePageDocumentListFragment.graphql";

import { CompliancePageDocumentListItem } from "./CompliancePageDocumentListItem";

const fragment = graphql`
  fragment CompliancePageDocumentListFragment on Organization {
    compliancePage: compliancePortal @required(action: THROW) {
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

  const { t } = useTranslation("organizations/compliance-page");

  const { compliancePage, documents } = useFragment<CompliancePageDocumentListFragment$key>(fragment, fragmentRef);
  const publishedDocuments = documents.edges.filter(({ node }) => node.currentPublishedMajor != null);

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{t("documentList.columns.name")}</Th>
            <Th>{t("documentList.columns.type")}</Th>
            <Th>{t("documentList.columns.alias")}</Th>
            <Th>{t("documentList.columns.visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {publishedDocuments.length === 0 && (
            <Tr>
              <Td colSpan={4} className="text-center text-txt-secondary">
                {t("documentList.empty")}
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
