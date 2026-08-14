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

import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDocumentList_compliancePortal$key } from "#/__generated__/core/CompliancePortalDocumentList_compliancePortal.graphql";
import type { CompliancePortalDocumentList_organization$key } from "#/__generated__/core/CompliancePortalDocumentList_organization.graphql";
import type {
  CompliancePortalDocumentListPaginationQuery,
  DocumentOrderField,
} from "#/__generated__/core/CompliancePortalDocumentListPaginationQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CompliancePortalDocumentListItem } from "./CompliancePortalDocumentListItem";

const compliancePortalFragment = graphql`
  fragment CompliancePortalDocumentList_compliancePortal on CompliancePortal {
    id
    ...CompliancePortalDocumentListItem_compliancePortal
  }
`;

const organizationFragment = graphql`
  fragment CompliancePortalDocumentList_organization on Organization
  @refetchable(queryName: "CompliancePortalDocumentListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    order: {
      type: "DocumentOrder"
      defaultValue: { field: TITLE, direction: ASC }
    }
    compliancePortalId: { type: "ID!" }
  ) {
    documents(
      first: $first
      after: $after
      before: $before
      last: $last
      orderBy: $order
      filter: { status: [ACTIVE], published: true }
    )
      @connection(
        key: "CompliancePortalDocumentList_documents"
        filters: ["orderBy", "filter"]
      ) {
      edges {
        node {
          id
          ...CompliancePortalDocumentListItem_document
            @arguments(compliancePortalId: $compliancePortalId)
        }
      }
    }
  }
`;

export function CompliancePortalDocumentList(props: {
  organizationKey: CompliancePortalDocumentList_organization$key;
  compliancePortalKey: CompliancePortalDocumentList_compliancePortal$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");
  const compliancePortal = useFragment(
    compliancePortalFragment,
    props.compliancePortalKey,
  );
  const pagination = usePaginationFragment<
    CompliancePortalDocumentListPaginationQuery,
    CompliancePortalDocumentList_organization$key
  >(organizationFragment, props.organizationKey);

  const documents = pagination.data.documents.edges.map(({ node }) => node);
  const refetch: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      compliancePortalId: compliancePortal.id,
      order: {
        direction: order.direction,
        field: order.field as DocumentOrderField,
      },
    });
  };

  return (
    <div className="space-y-2.5">
      <SortableTable
        {...pagination}
        refetch={refetch}
        initialOrder={{ field: "TITLE", direction: "ASC" }}
      >
        <Thead>
          <Tr>
            <Th className="w-24">{t("documentList.columns.displayed")}</Th>
            <SortableTh field="TITLE">
              {t("documentList.columns.name")}
            </SortableTh>
            <Th>{t("documentList.columns.type")}</Th>
            <Th>{t("documentList.columns.alias")}</Th>
            <Th>{t("documentList.columns.visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {documents.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {t("documentList.empty")}
              </Td>
            </Tr>
          )}
          {documents.map(document => (
            <CompliancePortalDocumentListItem
              key={document.id}
              documentKey={document}
              compliancePortalKey={compliancePortal}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
