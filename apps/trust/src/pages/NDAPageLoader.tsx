import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import { RelayProvider } from "#/providers/RelayProviders";

import type { NDAPageQuery } from "./__generated__/NDAPageQuery.graphql";
import { NDAPage, ndaPageQuery } from "./NDAPage";

function NDAPageQueryLoader() {
  const [queryRef, loadQuery]
    = useQueryLoader<NDAPageQuery>(ndaPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({});
    }
  });

  if (!queryRef) return null;

  return (
    <Suspense>
      <NDAPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function NDAPageLoader() {
  return (
    <RelayProvider>
      <NDAPageQueryLoader />
    </RelayProvider>
  );
}
