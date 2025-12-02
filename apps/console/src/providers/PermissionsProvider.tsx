import { useSuspenseQuery } from "@tanstack/react-query";
import { createContext, type PropsWithChildren } from "react";
import { useOrganizationId } from "/hooks/useOrganizationId";

enum Role {
  OWNER = "OWNER",
  ADMIN = "ADMIN",
  VIEWER = "VIEWER",
}

type PermissionsResponse = {
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

export function PermissionsProvider(props: PropsWithChildren) {
  const { children } = props;

  const organizationId = useOrganizationId();

  const { data } = useSuspenseQuery<PermissionsResponse>({
    queryKey: ["permissions", organizationId],
    queryFn: async () => {
      const response = await fetch(`/authz/${organizationId}/permissions`, { credentials: "include" });
      if (!response.ok) {
        throw new Error("Failed to fetch permissions");
      }
      return response.json() as Promise<PermissionsResponse>;
    },
  });

  const isAuthorized = (entity: string, action: string) => {
    return data.permissions[entity]?.[action] ?? false;
  }

  return (
    <PermissionsContext value={{ ...data, isAuthorized }}>
      {children}
    </PermissionsContext>
  );
}
