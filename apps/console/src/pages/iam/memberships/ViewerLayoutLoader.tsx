import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { ViewerLayoutQuery } from "#/__generated__/iam/ViewerLayoutQuery.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { ViewerLayout, viewerLayoutQuery } from "./ViewerLayout";
import { ViewerLayoutLoading } from "./ViewerLayoutLoading";

function ViewerLayoutQueryLoader() {
  const [queryRef, loadQuery]
    = useQueryLoader<ViewerLayoutQuery>(viewerLayoutQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) {
    return <ViewerLayoutLoading />;
  }

  return (
    <Suspense fallback={<ViewerLayoutLoading />}>
      <ViewerLayout queryRef={queryRef} />
    </Suspense>
  );
}

export default function ViewerLayoutLoader() {
  return (
    <IAMRelayProvider>
      <ViewerLayoutQueryLoader />
    </IAMRelayProvider>
  );
}
