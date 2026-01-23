import { CenteredLayoutSkeleton } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { MembershipsPageQuery } from "#/__generated__/iam/MembershipsPageQuery.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { MembershipsPage, membershipsPageQuery } from "./MembershipsPage";

function MembershipsPageQueryLoader() {
  const [queryRef, loadQuery]
    = useQueryLoader<MembershipsPageQuery>(membershipsPageQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) {
    return <CenteredLayoutSkeleton />;
  }

  return (
    <Suspense fallback={<CenteredLayoutSkeleton />}>
      <MembershipsPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function MembershipsPageLoader() {
  return (
    <IAMRelayProvider>
      <MembershipsPageQueryLoader />
    </IAMRelayProvider>
  );
}
