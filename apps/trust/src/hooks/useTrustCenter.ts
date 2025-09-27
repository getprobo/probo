import { TrustCenterContext } from "/providers/TrustCenterProvider";
import { useContext } from "react";

export function useTrustCenter(): {
  id: string;
  organization: { name: string };
} {
  const context = useContext(TrustCenterContext);
  if (!context) {
    throw new Error("useTrustCenter must be used within a TrustCenterProvider");
  }
  return context;
}
