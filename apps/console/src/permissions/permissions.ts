// Authorization system that checks permissions from the backend
// Permissions are fetched per organization from /authz/:organizationId/permissions
// Format: { permissions: { "Document": { "node": true, "updateDocument": true }, "Organization": { "createDocument": true } }, role: "ADMIN" }

export type EntityPermissions = Record<string, Record<string, boolean>>;

type PermissionsResponse = {
  permissions: EntityPermissions;
  role: string;
};

let cachedPermissions: EntityPermissions | null = null;
let cachedRole: string | null = null;
let cachePromise: Promise<PermissionsResponse> | null = null;
let currentOrganizationId: string | null = null;

/**
 * Fetch permissions for the current user's role in the organization
 */
function fetchPermissions(organizationId: string): Promise<PermissionsResponse> {
  if (cachedPermissions && cachedRole && currentOrganizationId === organizationId) {
    return Promise.resolve({ permissions: cachedPermissions, role: cachedRole });
  }

  if (cachePromise && currentOrganizationId === organizationId) {
    return cachePromise;
  }

  if (currentOrganizationId !== organizationId) {
    cachedPermissions = null;
    cachedRole = null;
    cachePromise = null;
    currentOrganizationId = organizationId;
  }

  const requestedOrgId = organizationId;

  cachePromise = fetch(`/authz/${encodeURIComponent(organizationId)}/permissions`, {
    credentials: 'include',
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Failed to fetch permissions: ${response.statusText}`);
      }
      return response.json();
    })
    .then((data: PermissionsResponse) => {
      if (currentOrganizationId === requestedOrgId) {
        cachedPermissions = data.permissions;
        cachedRole = data.role;
      }
      cachePromise = null;
      return data;
    })
    .catch((error) => {
      cachePromise = null;
      cachedPermissions = null;
      cachedRole = null;
      throw error;
    });

  return cachePromise;
}

/**
 * Check if the user has permission for an entity and action
 *
 * @param organizationId - The organization ID
 * @param entity - The entity name (e.g., "Document", "Organization", "Vendor")
 * @param action - The action/field name (e.g., "node", "updateDocument", "createDocument")
 * @returns true if the user has permission
 *
 * @example
 * isAuthorized(orgId, "Document", "get") // Check if user can query Document nodes
 * isAuthorized(orgId, "Document", "updateDocument") // Check if user can update documents
 * isAuthorized(orgId, "Organization", "createDocument") // Check if user can create documents
 */
export function isAuthorized(
  organizationId: string,
  entity: string,
  action: string
): boolean {
  if (!cachedPermissions || currentOrganizationId !== organizationId) {
    throw fetchPermissions(organizationId);
  }

  const entityPermissions = cachedPermissions[entity];
  if (!entityPermissions) {
    return false;
  }

  return entityPermissions[action] === true;
}

/**
 * Get the current user's role in the organization
 *
 * @param organizationId - The organization ID
 * @returns The user's role (e.g., "OWNER", "ADMIN", "VIEWER", "FULL")
 *
 * @example
 * getUserRole(orgId) // Returns "ADMIN"
 */
export function getUserRole(organizationId: string): string {
  if (!cachedRole || currentOrganizationId !== organizationId) {
    throw fetchPermissions(organizationId);
  }

  return cachedRole;
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

  if (currentRole === "OWNER" || currentRole === "FULL") {
    return ["OWNER", "ADMIN", "VIEWER"];
  }

  if (currentRole === "ADMIN") {
    return ["ADMIN", "VIEWER"];
  }

  return [];
}
