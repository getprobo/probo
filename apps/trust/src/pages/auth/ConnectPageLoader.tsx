import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";

import { RelayProvider } from "/providers/RelayProviders";

import { ConnectPage, connectPageQuery } from "./ConnectPage";
import type { ConnectPageQuery } from "./__generated__/ConnectPageQuery.graphql";

function ConnectPageLoader() {
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

export default function () {
  return (
    <RelayProvider>
      <ConnectPageLoader />
    </RelayProvider>
  );
}
