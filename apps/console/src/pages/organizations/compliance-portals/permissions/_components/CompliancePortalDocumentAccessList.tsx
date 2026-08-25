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

import type { CompliancePortalDocumentAccessInfo } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDocumentAccessList_access$key } from "#/__generated__/core/CompliancePortalDocumentAccessList_access.graphql";

import {
  documentAccessInfoFrom,
  rejectOrRevokeStatus,
  updateAccessInput,
} from "../_lib/documentAccessInfo";
import { useUpdateCompliancePortalAccess } from "../_lib/useUpdateCompliancePortalAccess";
import { documentAccessList } from "../variants";

import { CompliancePortalDocumentAccessBulkDialog } from "./CompliancePortalDocumentAccessBulkDialog";
import { CompliancePortalDocumentAccessListItem } from "./CompliancePortalDocumentAccessListItem";

const COLUMN_COUNT = 5;

const fragment = graphql`
  fragment CompliancePortalDocumentAccessList_access on CompliancePortalAccess {
    id
    availableDocumentAccesses(
      first: 100
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          ...documentAccessInfo_documentAccess
        }
      }
    }
  }
`;

interface CompliancePortalDocumentAccessListProps {
  accessKey: CompliancePortalDocumentAccessList_access$key;
  canUpdate: boolean;
}

export function CompliancePortalDocumentAccessList({
  accessKey,
  canUpdate,
}: CompliancePortalDocumentAccessListProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, heading, actions } = documentAccessList();
  const access = useFragment(fragment, accessKey);
  const documentAccesses = access.availableDocumentAccesses.edges.map(edge =>
    documentAccessInfoFrom(edge.node, t),
  );
  const [updateAccess, isUpdating] = useUpdateCompliancePortalAccess();

  const showGrantAll = canUpdate && documentAccesses.some(item => item.status !== "GRANTED");
  const showRejectOrRevokeAll = canUpdate
    && documentAccesses.some(item => item.status !== "REJECTED" && item.status !== "REVOKED");

  async function commit(updates: CompliancePortalDocumentAccessInfo[]) {
    await updateAccess({
      variables: { input: updateAccessInput(access.id, updates) },
    });
  }

  return (
    <section className={root()}>
      <div className={heading()}>
        <Heading level={3} size={3} weight="medium" highContrast>
          {t("documentAccessList.title")}
        </Heading>
        {canUpdate && (showGrantAll || showRejectOrRevokeAll) && (
          <div className={actions()}>
            {showGrantAll && (
              <CompliancePortalDocumentAccessBulkDialog
                action="grantAll"
                loading={isUpdating}
                onConfirm={() => commit(
                  documentAccesses
                    .filter(item => item.status !== "GRANTED")
                    .map(item => ({ ...item, status: "GRANTED" as const })),
                )}
              >
                <Button type="button" size={1} variant="soft" color="neutral" disabled={isUpdating}>
                  {t("documentAccessList.actions.grantAll")}
                </Button>
              </CompliancePortalDocumentAccessBulkDialog>
            )}
            {showRejectOrRevokeAll && (
              <CompliancePortalDocumentAccessBulkDialog
                action="rejectOrRevokeAll"
                loading={isUpdating}
                onConfirm={() => commit(
                  documentAccesses
                    .filter(item => item.status !== "REJECTED" && item.status !== "REVOKED")
                    .map(item => ({ ...item, status: rejectOrRevokeStatus(item.status) })),
                )}
              >
                <Button type="button" size={1} variant="soft" color="red" disabled={isUpdating}>
                  {t("documentAccessList.actions.rejectOrRevokeAll")}
                </Button>
              </CompliancePortalDocumentAccessBulkDialog>
            )}
          </div>
        )}
      </div>
      <Table size={2} variant="surface">
        <TableHeader>
          <TableRow>
            <TableColumnHeaderCell>{t("documentAccessList.columns.name")}</TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("documentAccessList.columns.type")}</TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("documentAccessList.columns.category")}</TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("documentAccessList.columns.access")}</TableColumnHeaderCell>
            <TableColumnHeaderCell />
          </TableRow>
        </TableHeader>
        <TableBody>
          {documentAccesses.length === 0
            ? (
                <TableRow>
                  <TableCell colSpan={COLUMN_COUNT}>
                    <Text size={2} color="faint">{t("documentAccessList.empty")}</Text>
                  </TableCell>
                </TableRow>
              )
            : documentAccesses.map(documentAccess => (
                <CompliancePortalDocumentAccessListItem
                  key={`${documentAccess.type}:${documentAccess.id}`}
                  documentAccess={documentAccess}
                  canUpdate={canUpdate}
                  disabled={isUpdating}
                  onGrant={item => void commit([{ ...item, status: "GRANTED" }])}
                  onRejectOrRevoke={item => void commit([item])}
                />
              ))}
        </TableBody>
      </Table>
    </section>
  );
}
