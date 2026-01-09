import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { CenteredLayoutSkeleton } from "@probo/ui";
import type { ViewerLayoutQuery } from "/__generated__/iam/ViewerLayoutQuery.graphql";
import { ViewerLayout, viewerLayoutQuery } from "./ViewerLayout";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

function ViewerLayoutLoader() {
  const [queryRef, loadQuery] =
    useQueryLoader<ViewerLayoutQuery>(viewerLayoutQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) {
    return <CenteredLayoutSkeleton />;
  }

  return (
    <Suspense fallback={<CenteredLayoutSkeleton />}>
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
