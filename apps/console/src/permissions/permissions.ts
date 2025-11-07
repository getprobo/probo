// Mirror of pkg/authz/permissions.go permission system
// Fetched from backend via REST API

export type Role = string;
export type Action = string;
export type EntityType = string;

type ActionPermissions = Record<string, boolean>;
type EntityPermissions = Record<string, ActionPermissions>;

let cachedPermissions: EntityPermissions | null = null;
let cachePromise: Promise<EntityPermissions> | null = null;
let currentOrganizationId: string | null = null;

/**
 * Fetch permissions for the current user in the given organization
 * This function is called automatically by isAuthorized() when needed
 */
function fetchPermissions(organizationId: string): Promise<EntityPermissions> {
  if (currentOrganizationId !== organizationId) {
    cachedPermissions = null;
    cachePromise = null;
    currentOrganizationId = organizationId;
  }

  if (cachedPermissions) {
    return Promise.resolve(cachedPermissions);
  }

  if (cachePromise) {
    return cachePromise;
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
    .then((data) => {
      // Only cache the result if we're still on the same organization
      if (currentOrganizationId === requestedOrgId) {
        cachedPermissions = data;
      }
      cachePromise = null;
      return data;
    })
    .catch((error) => {
      cachePromise = null;
      cachedPermissions = null;
      throw error;
    });

  return cachePromise;
}

/**
 * Clear the permissions cache (e.g., when switching organizations)
 */
export function clearPermissionsCache(): void {
  cachedPermissions = null;
  cachePromise = null;
  currentOrganizationId = null;
}

/**
 * Check if the user is authorized to perform an action on an entity type
 * This will automatically fetch permissions if not cached, using Suspense for loading state
 */
export function isAuthorized(
  organizationId: string,
  entityType: EntityType,
  action: Action
): boolean {
  if (!cachedPermissions || currentOrganizationId !== organizationId) {
    throw fetchPermissions(organizationId);
  }

  const actionPerms = cachedPermissions[entityType];
  if (!actionPerms) {
    return false;
  }

  return actionPerms[action] ?? false;
}
