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
