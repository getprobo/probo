import { Card, Spinner } from "@probo/ui";
import { useOutletContext } from "react-router";
import { TrustCenterDocumentsCard } from "/components/trustCenter/TrustCenterDocumentsCard";
import { useUpdateDocumentVisibilityMutation } from "/hooks/graph/TrustCenterDocumentGraph";
import type { TrustCenterDocumentsCardFragment$key } from "/components/trustCenter/__generated__/TrustCenterDocumentsCardFragment.graphql";

type ContextType = {
  organization: {
    documents?: {
      edges: Array<{
        node: TrustCenterDocumentsCardFragment$key;
      }>;
    };
  };
};

export default function TrustCenterDocumentsTab() {
  const { organization } = useOutletContext<ContextType>();
  const [updateDocumentVisibility, isUpdatingDocuments] = useUpdateDocumentVisibilityMutation();

  const documents = organization.documents?.edges?.map((edge) => edge.node) || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        {isUpdatingDocuments && <Spinner />}
      </div>
      <Card padded>
        <TrustCenterDocumentsCard
          documents={documents}
          params={{}}
          disabled={isUpdatingDocuments}
          onChangeVisibility={updateDocumentVisibility}
          variant="table"
        />
      </Card>
    </div>
  );
}
