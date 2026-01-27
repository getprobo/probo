import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentSignaturesPageQuery } from "#/__generated__/core/DocumentSignaturesPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentSignaturesPage, documentSignaturesPageQuery } from "./DocumentSignaturesPage";

function DocumentSignaturesPageQueryLoader() {
  const organizationId = useOrganizationId();
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const [queryRef, loadQuery] = useQueryLoader<DocumentSignaturesPageQuery>(documentSignaturesPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({
        organizationId,
        documentId: documentId,
        versionId: versionId ?? "",
        versionSpecified: !!versionId,
      });
    }
  });

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <DocumentSignaturesPage queryRef={queryRef} />;
}

export default function DocumentSignaturesPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<LinkCardSkeleton />}>
        <DocumentSignaturesPageQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
