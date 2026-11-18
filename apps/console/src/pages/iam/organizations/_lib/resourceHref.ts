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

/**
 * Detail URL for a GraphQL typename, or null when the console has no page.
 * Types whose page lives under a parent resource need `parentId`.
 */
export function resourceHref(
  organizationId: string,
  typename: string,
  id: string,
  parentId?: string,
): string | null {
  const base = `/organizations/${organizationId}`;

  switch (typename) {
    case "Control":
      return parentId
        ? `${base}/governance/frameworks/${parentId}/controls/${id}`
        : null;
    case "DataProtectionImpactAssessment":
      return parentId
        ? `${base}/privacy/processing-activities/${parentId}#dpia`
        : null;
    case "TransferImpactAssessment":
      return parentId
        ? `${base}/privacy/processing-activities/${parentId}#tia`
        : null;
    case "TreatmentPlan":
      return parentId
        ? `${base}/risk-management/risk-analyses/${parentId}/treatment-plans`
        : null;
    case "AiSystem":
      return `${base}/registries/ai-systems/${id}`;
    case "Asset":
      return `${base}/registries/assets/${id}`;
    case "Audit":
      return `${base}/governance/audits/${id}`;
    case "BusinessFunction":
      return `${base}/registries/business-functions/${id}`;
    case "Datum":
      return `${base}/registries/data/${id}`;
    case "Device":
      return `${base}/itam/devices/${id}`;
    case "Document":
      return `${base}/governance/documents/${id}`;
    case "Finding":
      return `${base}/governance/findings/${id}`;
    case "Framework":
      return `${base}/governance/frameworks/${id}`;
    case "InternalControl":
      return `${base}/governance/internal-controls/${id}`;
    case "Obligation":
      return `${base}/registries/obligations/${id}`;
    case "ProcessingActivity":
      return `${base}/privacy/processing-activities/${id}`;
    case "RightsRequest":
      return `${base}/privacy/rights-requests/${id}`;
    case "Risk":
      return `${base}/risk-management/risks/${id}`;
    case "RiskAnalysis":
      return `${base}/risk-management/risk-analyses/${id}`;
    case "StatementOfApplicability":
      return `${base}/governance/statements-of-applicability/${id}`;
    case "Task":
      return `${base}/governance/tasks/${id}`;
    case "ThirdParty":
      return `${base}/tprm/third-parties/${id}`;
    default:
      return null;
  }
}
