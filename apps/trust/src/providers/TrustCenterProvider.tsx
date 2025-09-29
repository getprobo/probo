import { createContext, type ReactNode } from "react";
import type { TrustGraphQuery$data } from "/queries/__generated__/TrustGraphQuery.graphql";

export const TrustCenterContext = createContext<
  TrustGraphQuery$data["trustCenterBySlug"] | null
>(null);

export const TrustCenterProvider = ({
  children,
  trustCenter,
}: {
  children: ReactNode;
  trustCenter: TrustGraphQuery$data["trustCenterBySlug"];
}) => {
  return (
    <TrustCenterContext.Provider value={trustCenter}>
      {children}
    </TrustCenterContext.Provider>
  );
};
