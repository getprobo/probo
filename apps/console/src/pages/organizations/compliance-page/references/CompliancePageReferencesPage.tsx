import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge } from "@probo/ui";
import { useRef } from "react";
import { ConnectionHandler, graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageReferenceListItemFragment$data } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import type { CompliancePageReferencesPageQuery } from "#/__generated__/core/CompliancePageReferencesPageQuery.graphql";
import { TrustCenterReferenceDialog, type TrustCenterReferenceDialogRef } from "#/components/trustCenter/TrustCenterReferenceDialog";

import { CompliancePageReferenceList } from "./_components/CompliancePageReferenceList";

export const compliancePageReferencesPageQuery = graphql`
  query CompliancePageReferencesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter @required(action: THROW) {
          id
          canCreateReference: permission(action: "core:trust-center-reference:create")
          ...CompliancePageReferenceListFragment
        }
      }
    }
  }
`;

export function CompliancePageReferencesPage(props: { queryRef: PreloadedQuery<CompliancePageReferencesPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const dialogRef = useRef<TrustCenterReferenceDialogRef>(null);

  const { organization } = usePreloadedQuery<CompliancePageReferencesPageQuery>(
    compliancePageReferencesPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const referencesConnectionId = ConnectionHandler.getConnectionID(
    organization.compliancePage.id,
    "CompliancePageReferenceList_references",
    { orderBy: { field: "RANK", direction: "ASC" } },
  );

  const handleCreate = () => {
    if (referencesConnectionId) {
      dialogRef.current?.openCreate(organization.compliancePage.id, referencesConnectionId);
    }
  };

  const handleEdit = (reference: CompliancePageReferenceListItemFragment$data) => {
    dialogRef.current?.openEdit(reference);
  };

  return (
    <div className="space-y-4">
      {organization.compliancePage?.id && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-base font-medium">{__("References")}</h2>
              <p className="text-sm text-txt-tertiary">
                {__("Showcase your customers and partners on your compliance page")}
              </p>
            </div>
            {organization.compliancePage?.canCreateReference && (
              <Button icon={IconPlusLarge} onClick={handleCreate}>
                {__("Add Reference")}
              </Button>
            )}
          </div>

          <CompliancePageReferenceList fragmentRef={organization.compliancePage} onEdit={handleEdit} />

          <TrustCenterReferenceDialog ref={dialogRef} />
        </div>
      )}
    </div>
  );
};
