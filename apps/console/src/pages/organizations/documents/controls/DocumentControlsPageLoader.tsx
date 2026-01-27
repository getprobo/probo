import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentControlsPageQuery } from "#/__generated__/core/DocumentControlsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentControlsPage, documentControlsPagePageQuery } from "./DocumentControlsPage";

function DocumentControlsPageQueryLoader() {
  const { documentId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const [queryRef, loadQuery] = useQueryLoader<DocumentControlsPageQuery>(documentControlsPagePageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ documentId });
    }
  });

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <DocumentControlsPage queryRef={queryRef} />;
}

export default function DocumentControlsPageLoader() {
  return (
    <CoreRelayProvider>
      <DocumentControlsPageQueryLoader />
    </CoreRelayProvider>
  );
}
