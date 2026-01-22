import { useTranslate } from "@probo/i18n";
import { Spinner } from "@probo/ui";
import { useOutletContext } from "react-router";

import type { TrustCenterGraphQuery$data } from "/__generated__/core/TrustCenterGraphQuery.graphql";
import { TrustCenterDocumentsCard } from "/components/trustCenter/TrustCenterDocumentsCard";
import { useUpdateDocumentVisibilityMutation } from "/hooks/graph/TrustCenterDocumentGraph";

export default function TrustCenterDocumentsTab() {
  const { __ } = useTranslate();
  const { organization } = useOutletContext<TrustCenterGraphQuery$data>();
  const [updateDocumentVisibility, isUpdatingDocuments]
    = useUpdateDocumentVisibilityMutation();

  const documents
    = organization.documents?.edges?.map(edge => edge.node) || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Documents")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage policies, procedures and compliance documents")}
          </p>
        </div>
        {isUpdatingDocuments && <Spinner />}
      </div>
      <TrustCenterDocumentsCard
        documents={documents}
        params={{}}
        disabled={isUpdatingDocuments}
        onChangeVisibility={updateDocumentVisibility}
        canUpdate={!!organization.trustCenter?.canUpdate}
      />
    </div>
  );
}
