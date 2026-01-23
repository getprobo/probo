import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import { RelayProvider } from "#/providers/RelayProviders";

import type { ConnectPageQuery } from "./__generated__/ConnectPageQuery.graphql";
import { ConnectPage, connectPageQuery } from "./ConnectPage";

function ConnectPageQueryLoader() {
  const [queryRef, loadQuery]
    = useQueryLoader<ConnectPageQuery>(connectPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({});
    }
  });

  if (!queryRef) return null;

  return (
    <Suspense>
      <ConnectPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function ConnectPageLoader() {
  return (
    <RelayProvider>
      <ConnectPageQueryLoader />
    </RelayProvider>
  );
}
