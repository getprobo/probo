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

import type { CompliancePortalDocumentList_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalDocumentList_compliancePortalFragment.graphql";

import { CompliancePortalDocumentListItem } from "./CompliancePortalDocumentListItem";
import { NewCompliancePortalDocumentDialog } from "./NewCompliancePortalDocumentDialog";

const compliancePortalFragment = graphql`
  fragment CompliancePortalDocumentList_compliancePortalFragment on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
    ...CompliancePortalDocumentListItem_compliancePortalFragment
    documents(first: 100)
      @connection(key: "CompliancePortalDocumentList_documents") {
      __id
      edges {
        node {
          id
          ...CompliancePortalDocumentListItem_catalogDocumentFragment
        }
      }
    }
  }
`;

export function CompliancePortalDocumentList(props: {
  compliancePortalRef: CompliancePortalDocumentList_compliancePortalFragment$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");

  const compliancePortal = useFragment(compliancePortalFragment, props.compliancePortalRef);

  return (
    <div className="space-y-[10px]">
      <div className="flex justify-end">
        <NewCompliancePortalDocumentDialog
          compliancePortalId={compliancePortal.id}
          connectionId={compliancePortal.documents.__id}
          disabled={!compliancePortal.canUpdate}
        />
      </div>
      <Table>
        <Thead>
          <Tr>
            <Th>{t("documentList.columns.name")}</Th>
            <Th>{t("documentList.columns.type")}</Th>
            <Th>{t("documentList.columns.alias")}</Th>
            <Th>{t("documentList.columns.visibility")}</Th>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {compliancePortal.documents.edges.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {t("documentList.empty")}
              </Td>
            </Tr>
          )}
          {compliancePortal.documents.edges.map(({ node: catalogDocument }) => (
            <CompliancePortalDocumentListItem
              key={catalogDocument.id}
              compliancePortalFragmentRef={compliancePortal}
              catalogDocumentFragmentRef={catalogDocument}
              documentsConnectionId={compliancePortal.documents.__id}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
