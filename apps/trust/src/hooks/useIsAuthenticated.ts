import { useContext } from "react";
import { AuthContext } from "/providers/AuthProvider";

export function useIsAuthenticated(): boolean {
  return useContext(AuthContext).isAuthenticated;
}
