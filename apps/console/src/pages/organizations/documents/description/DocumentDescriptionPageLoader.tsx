import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentDescriptionPage, documentDescriptionPageQuery } from "./DocumentDescriptionPage";

function DocumentDescriptionPageQueryLoader() {
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const [queryRef, loadQuery] = useQueryLoader<DocumentDescriptionPageQuery>(documentDescriptionPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({
        documentId: documentId,
        versionId: versionId ?? "",
        versionSpecified: !!versionId,
      });
    }
  });

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <DocumentDescriptionPage queryRef={queryRef} />;
}

export default function DocumentDescriptionPageLoader() {
  return (
    <CoreRelayProvider>
      <DocumentDescriptionPageQueryLoader />
    </CoreRelayProvider>
  );
}
