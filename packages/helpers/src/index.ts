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

export { objectKeys, objectEntries, cleanFormData } from "./object";
export { sprintf, faviconUrl, slugify } from "./string";
export {
  getCertificateProvisioningErrorMessage,
  getCustomDomainStatusBadgeLabel,
  getCustomDomainStatusBadgeVariant,
} from "./customDomain";
export {
  getTreatment,
  getRiskImpacts,
  getRiskLikelihoods,
  getSeverity,
} from "./risk";
export {
  withViewTransition,
  downloadFile,
  externalLinkProps,
  safeOpenUrl,
  focusSiblingElement,
} from "./dom";
export { times, groupBy, isEmpty } from "./array";
export { randomInt } from "./number";
export { getMeasureStateLabel, measureStates } from "./measure";
export { getRole, getRoles, peopleRoles } from "./people";
export { certificationCategoryLabel, certifications } from "./certifications";
export {
  getCountryName,
  getCountryOptions,
  getCountryLabel,
  countries,
  type CountryCode,
} from "./countries";
export {
  getDocumentTypeLabel,
  documentTypes,
  getDocumentClassificationLabel,
  documentClassifications,
  documentWriteModes,
  getDocumentWriteModeLabel,
} from "./documents";
export {
  controlMaturityLevels,
  getControlMaturityLevelLabel,
  type ControlMaturityLevel,
} from "./controls";
export { getAssetTypeVariant } from "./assets";
export {
  getAuditStateLabel,
  getAuditStateVariant,
  auditStates,
} from "./audits";
export {
  getStatusVariant,
  getStatusLabel,
  getStatusOptions,
} from "./registryStatus";
export {
  getObligationStatusVariant,
  getObligationStatusLabel,
  getObligationStatusOptions,
} from "./obligationStatus";
export {
    getObligationTypeLabel,
    getObligationTypeOptions,
} from "./obligationType";
export {
    getCompliancePortalVisibilityVariant,
    getCompliancePortalVisibilityLabel,
    getCompliancePortalVisibilityOptions,
    getCompliancePortalVisibilityVariant as getCompliancePageVisibilityVariant,
    getCompliancePortalVisibilityLabel as getCompliancePageVisibilityLabel,
    getCompliancePortalVisibilityOptions as getCompliancePageVisibilityOptions,
    compliancePortalVisibilities,
    compliancePortalVisibilities as compliancePageVisibilities,
    type CompliancePortalVisibility,
    type CompliancePortalVisibility as CompliancePageVisibility,
} from "./compliancePortalVisibility";
export { promisifyMutation } from "./relay";
export {
  acceptDocument,
  acceptSpreadsheet,
  acceptPresentation,
  acceptText,
  acceptImage,
  acceptData,
  acceptVideo,
  acceptAll,
} from "./fileAccept";
export {
  formatDatetime,
  formatDate,
  toDateInput,
  todayAsDateInput,
  parseDate,
} from "./date";
export {
  DURATION_UNITS,
  toMaxAgeSeconds,
  fromMaxAgeSeconds,
} from "./duration";
export { getTrackerTypeBadge, getTrackerSourceBadge } from "./tracker";
export { detectSocialName } from "./socialUrl";
export { formatError, type GraphQLError } from "./error";
export { Role, roles, getAssignableRoles } from "./roles";
export {
  getCompliancePortalDocumentAccessStatusBadgeVariant,
  getCompliancePortalDocumentAccessStatusLabel,
  getCompliancePortalDocumentAccessStatusBadgeVariant as getCompliancePageDocumentAccessStatusBadgeVariant,
  getCompliancePortalDocumentAccessStatusLabel as getCompliancePageDocumentAccessStatusLabel,
  type CompliancePortalDocumentAccessInfo,
  type CompliancePortalDocumentAccessInfo as CompliancePageDocumentAccessInfo,
} from "./compliancePortalDocumentAccess";
export {
  getRightsRequestTypeLabel,
  getRightsRequestTypeOptions,
  getRightsRequestStateVariant,
  getRightsRequestStateLabel,
  getRightsRequestStateOptions,
  rightsRequestTypes,
  rightsRequestStates,
  type RightsRequestType,
  type RightsRequestState,
} from "./rightsRequest";
