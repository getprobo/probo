import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { SignInPageQuery } from "#/__generated__/iam/SignInPageQuery.graphql";

import SignInPage, { signInPageQuery } from "./SignInPage";

function SignInPageQueryLoader() {
  const [queryRef, loadQuery] =
    useQueryLoader<SignInPageQuery>(signInPageQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) return null;

  return <SignInPage queryRef={queryRef} />;
}

export default function SignInPageLoader() {
  return <SignInPageQueryLoader />;
}
