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
import { Button, IconPlusLarge } from "@probo/ui";
import { useRef } from "react";
import { ConnectionHandler, graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageReferenceListItemFragment$data } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import type { CompliancePageReferencesPageQuery } from "#/__generated__/core/CompliancePageReferencesPageQuery.graphql";
import { CompliancePageReferenceDialog, type CompliancePageReferenceDialogRef } from "#/components/compliancePage/CompliancePageReferenceDialog";

import { CompliancePageReferenceList } from "./_components/CompliancePageReferenceList";

export const compliancePageReferencesPageQuery = graphql`
  query CompliancePageReferencesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: compliancePortal @required(action: THROW) {
          id
          canCreateReference: permission(action: "compliance-portal:portal-reference:create")
          ...CompliancePageReferenceListFragment
        }
      }
    }
  }
`;

export function CompliancePageReferencesPage(props: { queryRef: PreloadedQuery<CompliancePageReferencesPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const dialogRef = useRef<CompliancePageReferenceDialogRef>(null);

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

  const handleEdit = (reference: CompliancePageReferenceListItemFragment$data, rank: number) => {
    dialogRef.current?.openEdit(reference, rank);
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

          <CompliancePageReferenceDialog ref={dialogRef} />
        </div>
      )}
    </div>
  );
};
