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

import { Button, IconPlusLarge, useDialogRef } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePortalFilesPageQuery } from "#/__generated__/core/CompliancePortalFilesPageQuery.graphql";

import { CompliancePortalFileList } from "./_components/CompliancePortalFileList";
import { NewCompliancePortalFileDialog } from "./_components/NewCompliancePortalFileDialog";

export const compliancePortalFilesPageQuery = graphql`
  query CompliancePortalFilesPageQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        id
        canCreateCompliancePortalFile: permission(action: "compliance-portal:portal-file:create")
        ...CompliancePortalFileList_compliancePortalFragment
      }
    }
  }
`;

export default function CompliancePortalFilesPage(props: {
  queryRef: PreloadedQuery<CompliancePortalFilesPageQuery>;
}) {
  const { queryRef } = props;

  const { t } = useTranslation("organizations/compliance-portals");
  const createDialogRef = useDialogRef();

  const data = usePreloadedQuery<CompliancePortalFilesPageQuery>(compliancePortalFilesPageQuery, queryRef);
  if (data.compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  const compliancePortalId = data.compliancePortal.id;
  const filesConnectionId = ConnectionHandler.getConnectionID(
    compliancePortalId,
    "CompliancePortalFileList_compliancePortalFiles",
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{t("filesPage.title")}</h3>
          <p className="text-sm text-txt-tertiary">
            {t("filesPage.description")}
          </p>
        </div>
        {data.compliancePortal.canCreateCompliancePortalFile && (
          <Button
            icon={IconPlusLarge}
            onClick={() => createDialogRef.current?.open()}
          >
            {t("filesPage.actions.add")}
          </Button>
        )}
      </div>

      <CompliancePortalFileList
        compliancePortalRef={data.compliancePortal}
        filesConnectionId={filesConnectionId}
      />

      <NewCompliancePortalFileDialog
        connectionId={filesConnectionId}
        compliancePortalId={compliancePortalId}
        ref={createDialogRef}
      />
    </div>
  );
}
