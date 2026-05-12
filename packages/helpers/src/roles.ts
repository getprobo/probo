// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

export const Role = {
  OWNER: "OWNER",
  ADMIN: "ADMIN",
  VIEWER: "VIEWER",
  AUDITOR: "AUDITOR",
  EMPLOYEE: "EMPLOYEE",
} as const

export type Role = (typeof Role)[keyof typeof Role];
export const roles = [
  "OWNER",
  "ADMIN",
  "VIEWER",
  "AUDITOR",
  "EMPLOYEE",
] as const

export function getAssignableRoles(currentRole: Role): Role[] {
  if (currentRole === Role.OWNER) {
    return [Role.OWNER, Role.ADMIN, Role.VIEWER, Role.AUDITOR, Role.EMPLOYEE];
  }

  if (currentRole === Role.ADMIN) {
    return [Role.ADMIN, Role.VIEWER, Role.AUDITOR, Role.EMPLOYEE];
  }

  return [];
}
