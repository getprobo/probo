import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { UserPageQuery } from "#/__generated__/iam/UserPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { UserPage, userPageQuery } from "./UserPage";

function UserPageQueryLoader() {
  const { userId } = useParams();
  if (!userId) {
    throw new Error(":userId missing in route params");
  }
  const [queryRef, loadQuery] = useQueryLoader<UserPageQuery>(userPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ userId });
    }
  }, [loadQuery, queryRef, userId]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <UserPage queryRef={queryRef} />;
}

export default function UserPageLoader() {
  return (
    <IAMRelayProvider>
      <UserPageQueryLoader />
    </IAMRelayProvider>
  );
}
