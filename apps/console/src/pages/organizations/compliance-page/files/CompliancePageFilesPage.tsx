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

import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, useDialogRef } from "@probo/ui";
import { ConnectionHandler, graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageFilesPageQuery } from "#/__generated__/core/CompliancePageFilesPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CompliancePageFileList } from "./_components/CompliancePageFileList";
import { NewCompliancePageFileDialog } from "./_components/NewCompliancePageFileDialog";

export const compliancePageFilesPageQuery = graphql`
  query CompliancePageFilesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        canCreateCompliancePageFile: permission(action: "compliance-portal:portal-file:create")
      }
      ...CompliancePageFileListFragment
    }
  }
`;

export function CompliancePageFilesPage(props: {
  queryRef: PreloadedQuery<CompliancePageFilesPageQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const createDialogRef = useDialogRef();

  const { organization } = usePreloadedQuery<CompliancePageFilesPageQuery>(compliancePageFilesPageQuery, queryRef);

  const filesConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CompliancePageFileList_compliancePageFiles",
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Files")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Upload and manage files for your compliance page")}
          </p>
        </div>
        {organization.canCreateCompliancePageFile && (
          <Button
            icon={IconPlusLarge}
            onClick={() => createDialogRef.current?.open()}
          >
            {__("Add File")}
          </Button>
        )}
      </div>

      <CompliancePageFileList fragmentRef={organization} />

      <NewCompliancePageFileDialog
        connectionId={filesConnectionId}
        ref={createDialogRef}
      />
    </div>
  );
}
