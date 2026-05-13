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

/**
 * Type definitions for thirdParty data
 */

/**
 * Category types for thirdParties
 */
export type ThirdPartyCategory =
  | "ANALYTICS"
  | "CLOUD_MONITORING"
  | "CLOUD_PROVIDER"
  | "COLLABORATION"
  | "CUSTOMER_SUPPORT"
  | "DATA_STORAGE_AND_PROCESSING"
  | "DOCUMENT_MANAGEMENT"
  | "EMPLOYEE_MANAGEMENT"
  | "ENGINEERING"
  | "FINANCE"
  | "IDENTITY_PROVIDER"
  | "IT"
  | "MARKETING"
  | "OFFICE_OPERATIONS"
  | "OTHER"
  | "PASSWORD_MANAGEMENT"
  | "PRODUCT_AND_DESIGN"
  | "PROFESSIONAL_SERVICES"
  | "RECRUITING"
  | "SALES"
  | "SECURITY"
  | "VERSION_CONTROL";

/**
 * Represents a software thirdParty or service provider with associated metadata
 */
export interface ThirdParty {
  /** Display name of the thirdParty */
  name: string;

  /** Legal business name of the thirdParty */
  legalName?: string;

  /** Physical headquarters address */
  headquarterAddress?: string;

  /** ThirdParty's website URL */
  websiteUrl: string;

  /** URL to thirdParty's privacy policy */
  privacyPolicyUrl?: string;

  /** URL to thirdParty's terms of service */
  termsOfServiceUrl?: string;

  /** URL to thirdParty's service level agreement */
  serviceLevelAgreementUrl?: string;

  /** URL to service software agreement */
  serviceSoftwareAgreementUrl?: string;

  /** URL to thirdParty's data processing agreement */
  dataProcessingAgreementUrl?: string;

  /** URL to thirdParty's list of subprocessors */
  subprocessorsListUrl?: string;

  /** URL to thirdParty's business associate agreement */
  businessAssociateAgreementUrl?: string;

  /** Short description of the thirdParty/service */
  description?: string;

  /** Primary category for the thirdParty */
  category?: ThirdPartyCategory;

  /** Security or compliance certifications held by the thirdParty */
  certifications?: string[];

  /** URL to thirdParty's security page */
  securityPageUrl?: string;

  /** URL to thirdParty's trust page */
  trustPageUrl?: string;

  /** URL to thirdParty's status page */
  statusPageUrl?: string;

  /** Countries where the thirdParty is located */
  countries?: CountryCode[];
}

/**
 * Array of thirdParty data
 */
export type ThirdParties = ThirdParty[];

/**
 * Default export representing the entire thirdParty dataset
 */
declare const data: ThirdParties;

export default data;
