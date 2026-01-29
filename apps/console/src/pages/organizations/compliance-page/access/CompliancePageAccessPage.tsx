import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, Spinner, Table, Tbody, Td, Tr, useDialogRef } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePageAccessPageQuery } from "#/__generated__/core/CompliancePageAccessPageQuery.graphql";

import { CompliancePageAccessList } from "./_components/CompliancePageAccessList";
import { NewCompliancePageAccessDialog } from "./_components/NewCompliancePageAccessDialog";

export const compliancePageAccessPageQuery = graphql`
  query CompliancePageAccessPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter @required(action: THROW) {
          id
          canCreateAccess: permission(action: "core:trust-center-access:create")
          ...CompliancePageAccessListFragment
        }
      }
    }
  }
`;

export function CompliancePageAccessPage(props: { queryRef: PreloadedQuery<CompliancePageAccessPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [editingAccessId, setEditingAccessId] = useState<string | null>(null);

  const { organization } = usePreloadedQuery<CompliancePageAccessPageQuery>(compliancePageAccessPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.compliancePage.id,
    "CompliancePageAccessList_accesses",
    { orderBy: { field: "CREATED_AT", direction: "DESC" } },
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("External Access")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__(
              "Manage who can access your compliance page",
            )}
          </p>
        </div>
        {organization.compliancePage?.id
          && organization.compliancePage.canCreateAccess && (
          <Button
            icon={IconPlusLarge}
            onClick={() => {
              dialogRef.current?.open();
            }}
          >
            {__("Add Access")}
          </Button>
        )}
      </div>

      {organization.compliancePage
        ? (
            <CompliancePageAccessList
              editingAccessId={editingAccessId}
              fragmentRef={organization.compliancePage}
            />
          )
        : (
            <Table>
              <Tbody>
                <Tr>
                  <Td className="text-center text-txt-tertiary py-8">
                    <Spinner />
                  </Td>
                </Tr>
              </Tbody>
            </Table>
          )}

      <NewCompliancePageAccessDialog
        connectionId={connectionId}
        compliancePageId={organization.compliancePage.id}
        onCreate={setEditingAccessId}
        ref={dialogRef}
      />
    </div>
  );
}
