import { createContext } from "react";
import { Role } from "@probo/helpers";

export type PermissionsResponse = {
  permissions: Record<string, Record<string, boolean>>;
  role: Role;
};

type PermissionsContextType = {
  isAuthorized: (entity: string, action: string) => boolean;
} & PermissionsResponse;

export const PermissionsContext = createContext<PermissionsContextType>({
  permissions: {},
  role: Role.VIEWER,
  isAuthorized: () => false,
});
