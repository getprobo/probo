export const Role = {
  OWNER: "OWNER",
  ADMIN: "ADMIN",
  VIEWER: "VIEWER",
  AUDITOR: "AUDITOR",
  EMPLOYEE: "EMPLOYEE",
} as const

export type Role = (typeof Role)[keyof typeof Role];

export function getAssignableRoles(currentRole: Role): Role[] {
  if (currentRole === Role.OWNER) {
    return [Role.OWNER, Role.ADMIN, Role.VIEWER, Role.AUDITOR, Role.EMPLOYEE];
  }

  if (currentRole === Role.ADMIN) {
    return [Role.ADMIN, Role.VIEWER, Role.AUDITOR, Role.EMPLOYEE];
  }

  return [];
}
