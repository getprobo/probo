// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

type Translator = (s: string) => string;

export const Role = {
  OWNER: "OWNER",
  ADMIN: "ADMIN",
  VIEWER: "VIEWER",
  AUDITOR: "AUDITOR",
  EMPLOYEE: "EMPLOYEE",
  COMPLIANCE_MANAGER: "COMPLIANCE_MANAGER",
} as const

export type Role = (typeof Role)[keyof typeof Role];
export const roles = [
  "OWNER",
  "ADMIN",
  "VIEWER",
  "AUDITOR",
  "EMPLOYEE",
  "COMPLIANCE_MANAGER",
] as const

export function getAssignableRoles(currentRole: Role): Role[] {
  if (currentRole === Role.OWNER) {
    return [
      Role.OWNER,
      Role.ADMIN,
      Role.VIEWER,
      Role.AUDITOR,
      Role.EMPLOYEE,
      Role.COMPLIANCE_MANAGER,
    ];
  }

  if (currentRole === Role.ADMIN) {
    return [
      Role.ADMIN,
      Role.VIEWER,
      Role.AUDITOR,
      Role.EMPLOYEE,
      Role.COMPLIANCE_MANAGER,
    ];
  }

  return [];
}

export function getMembershipRoles(t: Translator) {
  return [
    {
      value: Role.OWNER,
      label: t("helpers.membershipRole.owner"),
    },
    {
      value: Role.ADMIN,
      label: t("helpers.membershipRole.admin"),
    },
    {
      value: Role.VIEWER,
      label: t("helpers.membershipRole.viewer"),
    },
    {
      value: Role.AUDITOR,
      label: t("helpers.membershipRole.auditor"),
    },
    {
      value: Role.EMPLOYEE,
      label: t("helpers.membershipRole.employee"),
    },
    {
      value: Role.COMPLIANCE_MANAGER,
      label: t("helpers.membershipRole.complianceManager"),
    },
  ] as const;
}

export function getMembershipRole(t: Translator, role?: string): string {
  switch (role) {
    case Role.OWNER:
      return t("helpers.membershipRole.owner");
    case Role.ADMIN:
      return t("helpers.membershipRole.admin");
    case Role.VIEWER:
      return t("helpers.membershipRole.viewer");
    case Role.AUDITOR:
      return t("helpers.membershipRole.auditor");
    case Role.EMPLOYEE:
      return t("helpers.membershipRole.employee");
    case Role.COMPLIANCE_MANAGER:
      return t("helpers.membershipRole.complianceManager");
    default:
      return t("helpers.common.unknown");
  }
}
