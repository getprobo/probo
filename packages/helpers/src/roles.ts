export enum Role {
  OWNER = "OWNER",
  ADMIN = "ADMIN",
  VIEWER = "VIEWER",
}

export function getAssignableRoles(currentRole: Role): string[] {
  if (!currentRole) return [];

  if (currentRole === "OWNER") {
    return ["OWNER", "ADMIN", "VIEWER"];
  }

  if (currentRole === "ADMIN") {
    return ["ADMIN", "VIEWER"];
  }

  return [];
}