import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import { RelayProvider } from "#/providers/RelayProviders";

import type { AuthLayoutQuery } from "./__generated__/AuthLayoutQuery.graphql";
import { AuthLayout, authLayoutQuery } from "./AuthLayout";

function AuthLayoutQueryLoader() {
  const [queryRef, loadQuery] = useQueryLoader<AuthLayoutQuery>(authLayoutQuery);

  useEffect(() => {
    if (!queryRef) {
      return loadQuery({});
    }
  });

  if (!queryRef) return null;

  return <AuthLayout queryRef={queryRef} />;
}

export default function AuthLayoutLoader() {
  return (
    <RelayProvider>
      <AuthLayoutQueryLoader />
    </RelayProvider>
  );
}
