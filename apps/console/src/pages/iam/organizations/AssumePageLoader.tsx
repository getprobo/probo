import { CenteredLayoutSkeleton } from "@probo/ui";
import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { AssumePageQuery } from "#/__generated__/iam/AssumePageQuery.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { AssumePage, assumePageQuery } from "./AssumePage";

function AssumePageQueryLoader() {
  const [queryRef, loadQuery] = useQueryLoader<AssumePageQuery>(assumePageQuery);

  useEffect(() => {
    loadQuery({}, { fetchPolicy: "network-only" });
  }, [loadQuery]);

  if (!queryRef) {
    return <CenteredLayoutSkeleton />;
  }

  return <AssumePage queryRef={queryRef} />;
}

export default function AssumePageLoader() {
  return (
    <IAMRelayProvider>
      <AssumePageQueryLoader />
    </IAMRelayProvider>
  );
}
