import { useEffect, type PropsWithChildren } from "react";
import { PermissionsContext } from "./NewPermissionsContext";
import {
  usePreloadedQuery,
  useQueryLoader,
  type PreloadedQuery,
} from "react-relay";
import type { GraphQLTaggedNode, OperationType } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";

type PermissionsQueryType = OperationType & {
  response: { viewer: Record<string, boolean> };
};

type PermissionsProviderLoaderProps = PropsWithChildren<{
  query: GraphQLTaggedNode;
}>;

export function PermissionsProviderLoader<T extends PermissionsQueryType>(
  props: PermissionsProviderLoaderProps,
) {
  const { children, query } = props;

  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<T>(query);

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    // While permissions are loading, consider there are no permissions
    return <PermissionsContext value={{}}>{children}</PermissionsContext>;
  }

  return (
    <PermissionsProvider query={query} queryRef={queryRef}>
      {children}
    </PermissionsProvider>
  );
}

type PermissionsProviderProps<T extends PermissionsQueryType> =
  PermissionsProviderLoaderProps & {
    queryRef: PreloadedQuery<T>;
  };

function PermissionsProvider<T extends PermissionsQueryType>(
  props: PermissionsProviderProps<T>,
) {
  const { children, query, queryRef } = props;

  const { viewer: permissions } = usePreloadedQuery<T>(query, queryRef);

  const handler = {
    get(target: Record<string, boolean>, prop: string) {
      if (!(prop in target)) {
        throw new Error(`Field ${prop} is not defined in IAM viewer node`);
      }

      return Reflect.get(target, prop);
    },
  };

  return (
    <PermissionsContext value={new Proxy(permissions, handler)}>
      {children}
    </PermissionsContext>
  );
}
