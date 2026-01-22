import { createContext, type ReactNode } from "react";

import type { TrustGraphCurrentQuery$data } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";

export const TrustCenterContext = createContext<
  TrustGraphCurrentQuery$data["currentTrustCenter"] | null
>(null);

export const TrustCenterProvider = ({
  children,
  trustCenter,
}: {
  children: ReactNode;
  trustCenter: TrustGraphCurrentQuery$data["currentTrustCenter"];
}) => {
  return (
    <TrustCenterContext.Provider value={trustCenter}>
      {children}
    </TrustCenterContext.Provider>
  );
};
