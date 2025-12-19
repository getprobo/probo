import { type PropsWithChildren } from "react";
import { PermissionsContext } from "./PermissionsContext";
import { Role } from "@probo/helpers";

export function PermissionsProvider(props: PropsWithChildren) {
  const { children } = props;

  // const organizationId = useOrganizationId();

  // const { data } = useSuspenseQuery<PermissionsResponse>({
  //   queryKey: ["permissions", organizationId],
  //   queryFn: async () => {
  //     const response = await fetch(`/authz/${organizationId}/permissions`, { credentials: "include" });
  //     if (!response.ok) {
  //       throw new Error("Failed to fetch permissions");
  //     }
  //     return response.json() as Promise<PermissionsResponse>;
  //   },
  // });

  // @ts-expect-error wip refactor
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const isAuthorized = (entity: string, action: string) => {
    // return data.permissions[entity]?.[action] ?? false;
    return true;
  };

  return (
    <PermissionsContext
      value={{ permissions: {}, role: Role.OWNER, isAuthorized }}
    >
      {children}
    </PermissionsContext>
  );

  // return (
  //   <PermissionsContext value={{ ...data, isAuthorized }}>
  //     {children}
  //   </PermissionsContext>
  // );
}
