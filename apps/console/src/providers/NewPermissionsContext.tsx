import { createContext } from "react";

export type PermissionsContextType = Record<string, boolean>;

export const PermissionsContext = createContext<PermissionsContextType>({});
