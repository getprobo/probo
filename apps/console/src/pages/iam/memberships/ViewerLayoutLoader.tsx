import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";

import type { ViewerLayoutQuery } from "/__generated__/iam/ViewerLayoutQuery.graphql";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

import { ViewerLayout, viewerLayoutQuery } from "./ViewerLayout";
import { ViewerLayoutLoading } from "./ViewerLayoutLoading";

function ViewerLayoutLoader() {
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

export default function () {
  return (
    <IAMRelayProvider>
      <ViewerLayoutLoader />
    </IAMRelayProvider>
  );
}
