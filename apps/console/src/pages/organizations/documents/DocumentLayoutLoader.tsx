import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentLayoutQuery } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentLayout, documentLayoutQuery } from "./DocumentLayout";

function DocumentLayoutQueryLoader() {
  const { documentId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }
  const [queryRef, loadQuery] = useQueryLoader<DocumentLayoutQuery>(documentLayoutQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ documentId });
    }
  });

  if (!queryRef) return <PageSkeleton />;

  return <DocumentLayout queryRef={queryRef} />;
}

export default function DocumentLayoutLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<PageSkeleton />}>
        <DocumentLayoutQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
