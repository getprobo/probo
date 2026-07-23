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

import type { CompliancePortalDocumentAccessStatus } from "@probo/coredata";
import type { CompliancePageDocumentAccessInfo } from "@probo/helpers";
import {
  getCompliancePageDocumentAccessStatusBadgeVariant,
  getCompliancePageDocumentAccessStatusLabel,
} from "@probo/helpers";
import { Badge, Button, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";

interface CompliancePageDocumentAccessListProps {
  documentAccesses: CompliancePageDocumentAccessInfo[];
  initialStatusByID: Record<string, CompliancePortalDocumentAccessStatus>;
  onGrantAll: () => void;
  onRejectOrRevokeAll: () => void;
  onUpdateStatus: (docAccess: CompliancePageDocumentAccessInfo, status: CompliancePortalDocumentAccessStatus) => void;
}

export function CompliancePageDocumentAccessList(props: CompliancePageDocumentAccessListProps) {
  const { documentAccesses, initialStatusByID, onGrantAll, onRejectOrRevokeAll, onUpdateStatus } = props;

  const { t } = useTranslation("organizations/compliance-page");

  const showGrantCTA = documentAccesses.some(da => da.status !== "GRANTED");
  const showRejectCTA = documentAccesses.some(da => da.status !== "REJECTED" && da.status !== "REVOKED");

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h4 className="font-medium text-txt-primary">
          {t("documentAccessList.title")}
        </h4>
        <div className="ml-auto flex items-center gap-2">
          {showGrantCTA
            && (
              <Button
                type="button"
                variant="quaternary"
                onClick={onGrantAll}
                className="text-xs h-7 min-h-7"
              >
                {t("documentAccessList.actions.grantAll")}
              </Button>
            )}
          {showRejectCTA
            && (
              <Button
                type="button"
                variant="danger"
                onClick={onRejectOrRevokeAll}
                className="text-xs h-7 min-h-7"
              >
                {t("documentAccessList.actions.rejectOrRevokeAll")}
              </Button>
            )}
        </div>
      </div>

      {documentAccesses.length > 0
        ? (
            <div className="bg-bg-secondary rounded-lg overflow-hidden">
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("documentAccessList.columns.name")}</Th>
                    <Th>{t("documentAccessList.columns.type")}</Th>
                    <Th>{t("documentAccessList.columns.category")}</Th>
                    <Th>
                      {t("documentAccessList.columns.access")}
                    </Th>
                    <Th></Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {documentAccesses.map((docAccess) => {
                    return (
                      <Tr key={docAccess.id}>
                        <Td>
                          <div className="font-medium text-txt-primary">
                            {docAccess.name}
                          </div>
                        </Td>
                        <Td>
                          <Badge variant={docAccess.variant}>
                            {docAccess.typeLabel}
                          </Badge>
                        </Td>
                        <Td>
                          <div className="text-txt-secondary">
                            {docAccess.category || "-"}
                          </div>
                        </Td>
                        <Td>
                          {(docAccess.persisted || docAccess.status !== "REQUESTED")
                            && (
                              <Badge variant={getCompliancePageDocumentAccessStatusBadgeVariant(docAccess.status)}>
                                {getCompliancePageDocumentAccessStatusLabel(docAccess.status, t)}
                              </Badge>
                            )}
                        </Td>
                        <Td className="flex justify-end gap-2">
                          {docAccess.status !== "GRANTED"
                            && (
                              <Button
                                type="button"
                                variant="quaternary"
                                onClick={() => onUpdateStatus(docAccess, "GRANTED")}
                                className="text-xs h-7 min-h-7"
                              >
                                {t("documentAccessList.actions.grant")}
                              </Button>
                            )}
                          {docAccess.status !== "REJECTED" && docAccess.status !== "REVOKED"
                            && (
                              <Button
                                type="button"
                                variant="danger"
                                onClick={() => onUpdateStatus(
                                  docAccess,
                                  docAccess.id && initialStatusByID[docAccess.id] === "GRANTED" ? "REVOKED" : "REJECTED",
                                )}
                                className="text-xs h-7 min-h-7"
                              >
                                {docAccess.id
                                  && initialStatusByID[docAccess.id] === "GRANTED"
                                  ? t("documentAccessList.actions.revoke")
                                  : t("documentAccessList.actions.reject")}
                              </Button>
                            )}
                        </Td>
                      </Tr>
                    );
                  })}
                </Tbody>
              </Table>
            </div>
          )
        : (
            <div className="text-center text-txt-tertiary py-8">
              {t("documentAccessList.empty")}
            </div>
          )}
    </div>
  );
}
