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

import { Button, IconPlusLarge } from "@probo/ui";
import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, useFragment } from "react-relay";

import type { CompliancePortalReferenceListItemFragment$data } from "#/__generated__/core/CompliancePortalReferenceListItemFragment.graphql";
import type { CompliancePortalReferencesSectionFragment$key } from "#/__generated__/core/CompliancePortalReferencesSectionFragment.graphql";
import { CompliancePortalReferenceDialog, type CompliancePortalReferenceDialogRef } from "#/components/compliancePortal/CompliancePortalReferenceDialog";

import { CompliancePortalReferenceList } from "./CompliancePortalReferenceList";

const fragment = graphql`
  fragment CompliancePortalReferencesSectionFragment on CompliancePortal {
    id
    canCreateReference: permission(action: "compliance-portal:portal-reference:create")
    ...CompliancePortalReferenceListFragment
  }
`;

export function CompliancePortalReferencesSection(props: {
  fragmentRef: CompliancePortalReferencesSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-portals");
  const dialogRef = useRef<CompliancePortalReferenceDialogRef>(null);

  const compliancePortal = useFragment(fragment, fragmentRef);

  const referencesConnectionId = ConnectionHandler.getConnectionID(
    compliancePortal.id,
    "CompliancePortalReferenceList_references",
    { orderBy: { field: "RANK", direction: "ASC" } },
  );

  const handleCreate = () => {
    if (referencesConnectionId) {
      dialogRef.current?.openCreate(compliancePortal.id, referencesConnectionId);
    }
  };

  const handleEdit = (reference: CompliancePortalReferenceListItemFragment$data, rank: number) => {
    dialogRef.current?.openEdit(reference, rank);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-medium">{t("referencesPage.title")}</h2>
          <p className="text-sm text-txt-tertiary">
            {t("referencesPage.description")}
          </p>
        </div>
        {compliancePortal.canCreateReference && (
          <Button icon={IconPlusLarge} onClick={handleCreate}>
            {t("referencesPage.actions.add")}
          </Button>
        )}
      </div>

      <CompliancePortalReferenceList fragmentRef={compliancePortal} onEdit={handleEdit} />

      <CompliancePortalReferenceDialog ref={dialogRef} />
    </div>
  );
}
