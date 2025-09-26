import { createContext, useMemo } from "react";

export const AuthContext = createContext({ isAuthenticated: false });

type Props = {
  children: React.ReactNode;
  isAuthenticated: boolean;
};

export function AuthProvider({ children, isAuthenticated }: Props) {
  const value = useMemo(() => ({ isAuthenticated }), [isAuthenticated]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
