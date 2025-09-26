import { useOutletContext } from "react-router";

export function useIsAuthenticated(): boolean {
  return useOutletContext<{ trustCenter: { isUserAuthenticated: boolean } }>()
    .trustCenter.isUserAuthenticated;
}
