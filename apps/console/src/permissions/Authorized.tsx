import { type ReactNode, useState, useEffect } from "react";
import { useParams } from "react-router";
import { isAuthorized } from "./permissions";

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
  const [hasAccess, setHasAccess] = useState<boolean | null>(null);

  useEffect(() => {
    if (!organizationId) {
      setHasAccess(false);
      return;
    }

    // Try to check authorization, catching promise throws
    try {
      const authorized = isAuthorized(organizationId, entity, action);
      setHasAccess(authorized);
    } catch (promise) {
      // If a promise is thrown (Suspense pattern), wait for it
      if (promise instanceof Promise) {
        promise
          .then(() => {
            // Permissions loaded, try again
            try {
              const authorized = isAuthorized(organizationId, entity, action);
              setHasAccess(authorized);
            } catch {
              setHasAccess(false);
            }
          })
          .catch(() => {
            setHasAccess(false);
          });
      } else {
        setHasAccess(false);
      }
    }
  }, [organizationId, entity, action]);

  if (!organizationId || hasAccess === null || hasAccess === false) {
    return fallback;
  }

  return children;
}
