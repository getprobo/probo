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

import type { CompliancePortalAuditList_compliancePortal$key } from "#/__generated__/core/CompliancePortalAuditList_compliancePortal.graphql";
import type { CompliancePortalAuditList_organization$key } from "#/__generated__/core/CompliancePortalAuditList_organization.graphql";
import type {
  AuditOrderField,
  CompliancePortalAuditListPaginationQuery,
} from "#/__generated__/core/CompliancePortalAuditListPaginationQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CompliancePortalAuditListItem } from "./CompliancePortalAuditListItem";

const compliancePortalFragment = graphql`
  fragment CompliancePortalAuditList_compliancePortal on CompliancePortal {
    id
    ...CompliancePortalAuditListItem_compliancePortal
  }
`;

const organizationFragment = graphql`
  fragment CompliancePortalAuditList_organization on Organization
  @refetchable(queryName: "CompliancePortalAuditListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    order: {
      type: "AuditOrder"
      defaultValue: { field: CREATED_AT, direction: DESC }
    }
    compliancePortalId: { type: "ID!" }
  ) {
    audits(
      first: $first
      after: $after
      before: $before
      last: $last
      orderBy: $order
    )
      @connection(
        key: "CompliancePortalAuditList_audits"
        filters: ["orderBy"]
      ) {
      edges {
        node {
          id
          ...CompliancePortalAuditListItem_audit
            @arguments(compliancePortalId: $compliancePortalId)
        }
      }
    }
  }
`;

export function CompliancePortalAuditList(props: {
  organizationKey: CompliancePortalAuditList_organization$key;
  compliancePortalKey: CompliancePortalAuditList_compliancePortal$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");
  const compliancePortal = useFragment(
    compliancePortalFragment,
    props.compliancePortalKey,
  );
  const pagination = usePaginationFragment<
    CompliancePortalAuditListPaginationQuery,
    CompliancePortalAuditList_organization$key
  >(organizationFragment, props.organizationKey);

  const audits = pagination.data.audits.edges.map(({ node }) => node);
  const refetch: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      compliancePortalId: compliancePortal.id,
      order: {
        direction: order.direction,
        field: order.field as AuditOrderField,
      },
    });
  };

  return (
    <div className="space-y-2.5">
      <SortableTable
        {...pagination}
        refetch={refetch}
        initialOrder={{ field: "CREATED_AT", direction: "DESC" }}
      >
        <Thead>
          <Tr>
            <Th className="w-24">{t("auditList.columns.displayed")}</Th>
            <Th>{t("auditList.columns.framework")}</Th>
            <Th>{t("auditList.columns.name")}</Th>
            <SortableTh field="VALID_UNTIL">
              {t("auditList.columns.validUntil")}
            </SortableTh>
            <SortableTh field="STATE">
              {t("auditList.columns.state")}
            </SortableTh>
            <Th>{t("auditList.columns.visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {audits.length === 0 && (
            <Tr>
              <Td colSpan={6} className="text-center text-txt-secondary">
                {t("auditList.empty")}
              </Td>
            </Tr>
          )}
          {audits.map(audit => (
            <CompliancePortalAuditListItem
              key={audit.id}
              auditKey={audit}
              compliancePortalKey={compliancePortal}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
