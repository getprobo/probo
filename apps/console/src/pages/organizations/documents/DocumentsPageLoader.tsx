import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { DocumentsPageQuery } from "#/__generated__/core/DocumentsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import DocumentsPage, { documentsPageQuery } from "./DocumentsPage";

function DocumentsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<DocumentsPageQuery>(documentsPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <PageSkeleton />;

  return <DocumentsPage queryRef={queryRef} />;
}

export default function DocumentsPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<PageSkeleton />}>
        <DocumentsPageQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
