import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { PersonPageQuery } from "#/__generated__/iam/PersonPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { PersonPage, personPageQuery } from "./PersonPage";

function PersonPageQueryLoader() {
  const { personId } = useParams();
  if (!personId) {
    throw new Error(":personId missing in route params");
  }
  const [queryRef, loadQuery] = useQueryLoader<PersonPageQuery>(personPageQuery);

  useEffect(() => {
    loadQuery({ personId });
  }, [personId, loadQuery]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <PersonPage queryRef={queryRef} />;
}

export default function PersonPageLoader() {
  return (
    <IAMRelayProvider>
      <PersonPageQueryLoader />
    </IAMRelayProvider>
  );
}
