import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge } from "@probo/ui";
import { useRef } from "react";
import { ConnectionHandler, graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageBadgeListItemFragment$data } from "#/__generated__/core/CompliancePageBadgeListItemFragment.graphql";
import type { CompliancePageReferenceListItemFragment$data } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import type { CompliancePageReferencesPageQuery } from "#/__generated__/core/CompliancePageReferencesPageQuery.graphql";
import { ComplianceBadgeDialog, type ComplianceBadgeDialogRef } from "#/components/trustCenter/ComplianceBadgeDialog";
import { TrustCenterReferenceDialog, type TrustCenterReferenceDialogRef } from "#/components/trustCenter/TrustCenterReferenceDialog";

import { CompliancePageBadgeList } from "./_components/CompliancePageBadgeList";
import { CompliancePageReferenceList } from "./_components/CompliancePageReferenceList";

export const compliancePageReferencesPageQuery = graphql`
  query CompliancePageReferencesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter @required(action: THROW) {
          id
          canCreateReference: permission(action: "core:trust-center-reference:create")
          canCreateComplianceBadge
          ...CompliancePageReferenceListFragment
          ...CompliancePageBadgeListFragment
        }
      }
    }
  }
`;

export function CompliancePageReferencesPage(props: { queryRef: PreloadedQuery<CompliancePageReferencesPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const referenceDialogRef = useRef<TrustCenterReferenceDialogRef>(null);
  const badgeDialogRef = useRef<ComplianceBadgeDialogRef>(null);

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

  const badgesConnectionId = ConnectionHandler.getConnectionID(
    organization.compliancePage.id,
    "CompliancePageBadgeList_complianceBadges",
    { orderBy: { field: "RANK", direction: "ASC" } },
  );

  const handleCreateReference = () => {
    if (referencesConnectionId) {
      referenceDialogRef.current?.openCreate(organization.compliancePage.id, referencesConnectionId);
    }
  };

  const handleEditReference = (reference: CompliancePageReferenceListItemFragment$data) => {
    referenceDialogRef.current?.openEdit(reference);
  };

  const handleCreateBadge = () => {
    if (badgesConnectionId) {
      badgeDialogRef.current?.openCreate(organization.compliancePage.id, badgesConnectionId);
    }
  };

  const handleEditBadge = (badge: CompliancePageBadgeListItemFragment$data) => {
    badgeDialogRef.current?.openEdit(badge);
  };

  return (
    <div className="space-y-8">
      {/* Badges Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-base font-semibold">{__("Badges")}</h2>
            <p className="text-sm text-txt-tertiary">
              {__("Display compliance and certification badges on your public page. Badges replace the frameworks section when present.")}
            </p>
          </div>
          {organization.compliancePage?.canCreateComplianceBadge && (
            <Button icon={IconPlusLarge} onClick={handleCreateBadge}>
              {__("Add Badge")}
            </Button>
          )}
        </div>

        <CompliancePageBadgeList fragmentRef={organization.compliancePage} onEdit={handleEditBadge} />
      </div>

      {/* References Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-base font-semibold">{__("References")}</h2>
            <p className="text-sm text-txt-tertiary">
              {__("Showcase your customers and partners on your compliance page")}
            </p>
          </div>
          {organization.compliancePage?.canCreateReference && (
            <Button icon={IconPlusLarge} onClick={handleCreateReference}>
              {__("Add Reference")}
            </Button>
          )}
        </div>

        <CompliancePageReferenceList fragmentRef={organization.compliancePage} onEdit={handleEditReference} />
      </div>

      <TrustCenterReferenceDialog ref={referenceDialogRef} />
      <ComplianceBadgeDialog ref={badgeDialogRef} />
    </div>
  );
}
