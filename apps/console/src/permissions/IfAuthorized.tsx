import { type ReactNode } from "react";
import { useParams } from "react-router";
import { isAuthorized, type EntityType, type Action } from "./permissions";

type Props = {
  entity: EntityType;
  action: Action;
  children: ReactNode;
  fallback?: ReactNode;
};

/**
 * Conditionally render children based on authorization check
 * Automatically fetches permissions using Suspense if not cached
 *
 * @example
 * <IfAuthorized entity="Document" action="delete">
 *   <DeleteButton />
 * </IfAuthorized>
 */
export function IfAuthorized({
  entity,
  action,
  children,
  fallback = null,
}: Props) {
  const { organizationId } = useParams();

  if (!organizationId) {
    return fallback;
  }

  const hasAccess = isAuthorized(organizationId, entity, action);

  return hasAccess ? children : fallback;
}
