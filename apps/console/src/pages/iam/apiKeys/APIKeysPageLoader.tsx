import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { CenteredLayoutSkeleton } from "@probo/ui";
import { APIKeysPage, apiKeysPageQuery } from "./APIKeysPage";
import type { APIKeysPageQuery } from "./__generated__/APIKeysPageQuery.graphql";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

function APIKeysPageLoaderInner() {
  const [queryRef, loadQuery] =
    useQueryLoader<APIKeysPageQuery>(apiKeysPageQuery);

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
