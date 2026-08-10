// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

export type AiSystemStatus = "ACTIVE" | "IN_DEVELOPMENT" | "DECOMMISSIONED";

export type AiSystemRiskClassification = "HIGH_RISK" | "LIMITED" | "MINIMAL" | "GPAI";

export type AiSystemCompanyRole = "PROVIDER" | "DEPLOYER" | "USER" | "DEVELOPER";

type BadgeVariant
  = "success" | "warning" | "danger" | "info" | "neutral" | "outline" | "highlight";

export const AI_SYSTEM_STATUSES: AiSystemStatus[] = [
  "ACTIVE",
  "IN_DEVELOPMENT",
  "DECOMMISSIONED",
];

export const AI_SYSTEM_RISK_CLASSIFICATIONS: AiSystemRiskClassification[] = [
  "HIGH_RISK",
  "LIMITED",
  "MINIMAL",
  "GPAI",
];

export const AI_SYSTEM_COMPANY_ROLES: AiSystemCompanyRole[] = [
  "PROVIDER",
  "DEPLOYER",
  "USER",
  "DEVELOPER",
];

export function getStatusLabel(
  status: AiSystemStatus,
  t: (key: string) => string,
  prefix: string,
): string {
  switch (status) {
    case "ACTIVE":
      return t(`${prefix}.status.active`);
    case "IN_DEVELOPMENT":
      return t(`${prefix}.status.inDevelopment`);
    case "DECOMMISSIONED":
      return t(`${prefix}.status.decommissioned`);
    default:
      return status;
  }
}

export function getStatusVariant(status: AiSystemStatus): BadgeVariant {
  switch (status) {
    case "ACTIVE":
      return "success";
    case "IN_DEVELOPMENT":
      return "warning";
    case "DECOMMISSIONED":
      return "neutral";
    default:
      return "neutral";
  }
}

export function getRiskClassificationLabel(
  classification: AiSystemRiskClassification,
  t: (key: string) => string,
  prefix: string,
): string {
  switch (classification) {
    case "HIGH_RISK":
      return t(`${prefix}.riskClassification.highRisk`);
    case "LIMITED":
      return t(`${prefix}.riskClassification.limited`);
    case "MINIMAL":
      return t(`${prefix}.riskClassification.minimal`);
    case "GPAI":
      return t(`${prefix}.riskClassification.gpai`);
    default:
      return classification;
  }
}

export function getRiskClassificationVariant(
  classification: AiSystemRiskClassification,
): BadgeVariant {
  switch (classification) {
    case "HIGH_RISK":
      return "danger";
    case "LIMITED":
      return "warning";
    case "MINIMAL":
      return "success";
    case "GPAI":
      return "info";
    default:
      return "neutral";
  }
}

export function getCompanyRoleLabel(
  role: AiSystemCompanyRole,
  t: (key: string) => string,
  prefix: string,
): string {
  switch (role) {
    case "PROVIDER":
      return t(`${prefix}.companyRoles.provider`);
    case "DEPLOYER":
      return t(`${prefix}.companyRoles.deployer`);
    case "USER":
      return t(`${prefix}.companyRoles.user`);
    case "DEVELOPER":
      return t(`${prefix}.companyRoles.developer`);
    default:
      return role;
  }
}

export const AiSystemsConnectionKey = "AiSystemsPage_aiSystems";

export const emptyAiSystemFilter = {
  status: null,
  riskClassification: null,
};

export type AiSystemListFilter = {
  status: AiSystemStatus | null;
  riskClassification: AiSystemRiskClassification | null;
};

export function aiSystemListConnectionFilters(aiSystem: {
  status?: string | null;
  riskClassification?: string | null;
}): AiSystemListFilter[] {
  const status = (aiSystem.status ?? null) as AiSystemStatus | null;
  const riskClassification = (aiSystem.riskClassification ?? null) as
    | AiSystemRiskClassification
    | null;

  const candidates: AiSystemListFilter[] = [
    emptyAiSystemFilter,
    { status, riskClassification: null },
    { status: null, riskClassification },
    { status, riskClassification },
  ];

  const seen = new Set<string>();
  const unique: AiSystemListFilter[] = [];

  for (const filter of candidates) {
    const key = `${filter.status ?? ""}:${filter.riskClassification ?? ""}`;
    if (seen.has(key)) {
      continue;
    }

    seen.add(key);
    unique.push(filter);
  }

  return unique;
}
