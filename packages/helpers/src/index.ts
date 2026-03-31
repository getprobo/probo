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

export { objectKeys, objectEntries, cleanFormData } from "./object";
export { sprintf, faviconUrl, slugify } from "./string";
export {
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
} from "./documents";
export { getAssetTypeVariant } from "./assets";
export {
  getSnapshotTypeLabel,
  getSnapshotTypeUrlPath,
  snapshotTypes,
  validateSnapshotConsistency,
} from "./snapshots";
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
    getTrustCenterVisibilityVariant,
    getTrustCenterVisibilityLabel,
    getTrustCenterVisibilityOptions,
    trustCenterVisibilities,
    type TrustCenterVisibility,
} from "./trustCenterVisibility";
export { promisifyMutation } from "./relay";
export { fileType, fileSize } from "./file";
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
  formatDuration,
  parseDate,
} from "./date";
export { getTrustCenterUrl } from "./trustCenter";
export { detectSocialName } from "./socialUrl";
export { formatError, type GraphQLError } from "./error";
export { Role, roles, getAssignableRoles } from "./roles";
export {
  getTrustCenterDocumentAccessStatusBadgeVariant,
  getTrustCenterDocumentAccessStatusLabel,
  type TrustCenterDocumentAccessInfo,
} from "./trustCenterDocumentAccess";
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
