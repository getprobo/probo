import { useEffect, useState } from 'react';

// Authorization system that checks permissions from the backend
// Permissions are fetched per organization from /authz/:organizationId/permissions
// Format: { permissions: { "Document": { "node": true, "updateDocument": true }, "Organization": { "createDocument": true } }, role: "ADMIN" }

export type EntityPermissions = Record<string, Record<string, boolean>>;

type PermissionsResponse = {
  permissions: EntityPermissions;
  role: string;
};

type PermissionsCache = {
  [organizationId: string]: {
    permissions: EntityPermissions;
    role: string;
  };
};

const cache: PermissionsCache = {};
const pendingRequests: Map<string, Promise<PermissionsResponse>> = new Map();

/**
 * Fetch permissions for the current user's role in the organization
 */
export function fetchPermissions(organizationId: string): Promise<PermissionsResponse> {
  if (cache[organizationId]) {
    return Promise.resolve(cache[organizationId]);
  }

  const pending = pendingRequests.get(organizationId);
  if (pending) {
    return pending;
  }

  const promise = fetch(`/authz/${encodeURIComponent(organizationId)}/permissions`, {
    credentials: 'include',
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Failed to fetch permissions: ${response.statusText}`);
      }
      return response.json();
    })
    .then((data: PermissionsResponse) => {
      if (!data || typeof data.permissions !== 'object' || !data.role) {
        throw new Error('Invalid permissions response structure');
      }

      cache[organizationId] = {
        permissions: data.permissions,
        role: data.role,
      };
      pendingRequests.delete(organizationId);
      return data;
    })
    .catch((error) => {
      pendingRequests.delete(organizationId);
      throw error;
    });

  pendingRequests.set(organizationId, promise);
  return promise;
}

/**
 * React hook to fetch and manage permissions for an organization
 *
 * @param organizationId - The organization ID
 * @returns Object with loading, error, permissions, role, and helper functions
 *
 * @example
 * const { loading, error, permissions, role, isAuthorized, getAssignableRoles } = usePermissions(orgId);
 *
 * if (loading) return <Spinner />;
 * if (error) return <ErrorMessage />;
 * if (isAuthorized("Document", "updateDocument")) { ... }
 */
export function usePermissions(organizationId: string) {
  const [state, setState] = useState<{
    loading: boolean;
    error: Error | null;
    permissions: EntityPermissions | null;
    role: string | null;
  }>({
    loading: true,
    error: null,
    permissions: null,
    role: null,
  });

  useEffect(() => {
    setState({ loading: true, error: null, permissions: null, role: null });

    fetchPermissions(organizationId)
      .then((data) => {
        setState({
          loading: false,
          error: null,
          permissions: data.permissions,
          role: data.role,
        });
      })
      .catch((error) => {
        setState({
          loading: false,
          error,
          permissions: null,
          role: null,
        });
      });
  }, [organizationId]);

  const checkAuthorized = (entity: string, action: string): boolean => {
    if (!state.permissions) return false;

    const entityPermissions = state.permissions[entity];
    if (!entityPermissions) return false;

    return entityPermissions[action] === true;
  };

  const getAssignableRolesList = (): string[] => {
    if (!state.role) return [];

    if (state.role === "OWNER" || state.role === "FULL") {
      return ["OWNER", "ADMIN", "VIEWER"];
    }

    if (state.role === "ADMIN") {
      return ["ADMIN", "VIEWER"];
    }

    return [];
  };

  return {
    loading: state.loading,
    error: state.error,
    permissions: state.permissions,
    role: state.role,
    isAuthorized: checkAuthorized,
    getAssignableRoles: getAssignableRolesList,
  };
}

/**
 * Check if the user has permission for an entity and action (synchronous)
 * Returns false if permissions are not loaded yet
 *
 * @param organizationId - The organization ID
 * @param entity - The entity name (e.g., "Document", "Organization", "Vendor")
 * @param action - The action/field name (e.g., "node", "updateDocument", "createDocument")
 * @returns true if the user has permission, false otherwise
 *
 * @example
 * isAuthorized(orgId, "Document", "get")
 * isAuthorized(orgId, "Document", "updateDocument")
 * isAuthorized(orgId, "Organization", "createDocument")
 */
export function isAuthorized(
  organizationId: string,
  entity: string,
  action: string
): boolean {
  const cached = cache[organizationId];
  if (!cached) return false;

  const entityPermissions = cached.permissions[entity];
  if (!entityPermissions) return false;

  return entityPermissions[action] === true;
}

/**
 * Get the current user's role in the organization
 * Returns empty string if not loaded yet
 *
 * @param organizationId - The organization ID
 * @returns The user's role (e.g., "OWNER", "ADMIN", "VIEWER", "FULL")
 *
 * @example
 * getUserRole(orgId) // Returns "ADMIN"
 */
export function getUserRole(organizationId: string): string {
  const cached = cache[organizationId];
  return cached?.role || "";
}

/**
 * Get available roles that the current user can assign
 * Based on the rule: OWNER and FULL can assign any role, ADMIN can assign ADMIN and VIEWER but not OWNER
 *
 * @param organizationId - The organization ID
 * @returns Array of roles that can be assigned
 */
export function getAssignableRoles(organizationId: string): string[] {
  const currentRole = getUserRole(organizationId);
  if (!currentRole) return [];

  if (currentRole === "OWNER" || currentRole === "FULL") {
    return ["OWNER", "ADMIN", "VIEWER"];
  }

  if (currentRole === "ADMIN") {
    return ["ADMIN", "VIEWER"];
  }

  return [];
}

/**
 * Check if the current user has a specific role in the organization
 * Fetches permissions if not already cached
 *
 * @param organizationId - The organization ID
 * @param role - The role to check (e.g., "EMPLOYEE", "OWNER", "ADMIN", "VIEWER", "FULL")
 * @returns Promise that resolves to true if the user has the role, false otherwise
 *
 * @example
 * const isEmployee = await isRole(orgId, "EMPLOYEE");
 * if (isEmployee) { ... }
 */
export async function isRole(organizationId: string, role: string): Promise<boolean> {
  try {
    const data = await fetchPermissions(organizationId);
    return data.role === role;
  } catch (error) {
    console.error('Failed to fetch permissions for role check:', error);
    return false;
  }
}
