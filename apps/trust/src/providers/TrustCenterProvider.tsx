import { createContext } from "react";
import type { TrustGraphQuery$data } from "/queries/__generated__/TrustGraphQuery.graphql";

export const TrustCenterContext = createContext(
  {} as TrustGraphQuery$data["trustCenterBySlug"]
);

export const TrustCenterProvider = ({
  children,
  trustCenter,
}: {
  children: React.ReactNode;
  trustCenter: TrustGraphQuery$data["trustCenterBySlug"];
}) => {
  return (
    <TrustCenterContext.Provider value={trustCenter}>
      {children}
    </TrustCenterContext.Provider>
  );
};
