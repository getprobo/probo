import { CenteredLayoutSkeleton } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { APIKeysPageQuery } from "#/__generated__/iam/APIKeysPageQuery.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { APIKeysPage, apiKeysPageQuery } from "./APIKeysPage";

function APIKeysPageLoaderInner() {
  const [queryRef, loadQuery]
    = useQueryLoader<APIKeysPageQuery>(apiKeysPageQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) {
    return <CenteredLayoutSkeleton />;
  }

  return (
    <Suspense fallback={<CenteredLayoutSkeleton />}>
      <APIKeysPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function APIKeysPageLoader() {
  return (
    <IAMRelayProvider>
      <APIKeysPageLoaderInner />
    </IAMRelayProvider>
  );
}
