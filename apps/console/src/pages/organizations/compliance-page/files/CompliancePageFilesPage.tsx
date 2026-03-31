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
        canCreateTrustCenterFile: permission(action: "core:trust-center-file:create")
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
    "CompliancePageFileList_trustCenterFiles",
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
        {organization.canCreateTrustCenterFile && (
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
