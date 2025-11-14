import { type ReactNode } from "react";
import { useParams } from "react-router";
import { usePermissions } from "./permissions";

type Props = {
  entity: string;
  action: string;
  children: ReactNode;
  fallback?: ReactNode;
};

/**
 * Conditionally render children based on authorization check
 * Automatically fetches permissions if not cached
 *
 * @param entity - The entity name (e.g., "Document", "Organization", "Vendor")
 * @param action - The action/field name (e.g., "get", "createDocument", "updateVendor")
 *
 * @example
 * <Authorized entity="Organization" action="createDocument">
 *   <CreateButton />
 * </Authorized>
 *
 * @example
 * <Authorized entity="Document" action="get">
 *   <DocumentViewer />
 * </Authorized>
 *
 * @example
 * <Authorized entity="Vendor" action="updateVendor">
 *   <EditButton />
 * </Authorized>
 */
export function Authorized({
  entity,
  action,
  children,
  fallback = null,
}: Props) {
  const { organizationId } = useParams();
  const { loading, error, isAuthorized } = usePermissions(organizationId || "");

  if (!organizationId || loading || error || !isAuthorized(entity, action)) {
    return fallback;
  }

  return children;
}
