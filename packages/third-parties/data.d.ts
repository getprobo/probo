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
 * Type definitions for third party data
 */

/**
 * Category types for third partys
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
 * Represents a software third party or service provider with associated metadata
 */
export interface ThirdParty {
  /** Display name of the third party */
  name: string;

  /** Legal business name of the third party */
  legalName?: string;

  /** Physical headquarters address */
  headquarterAddress?: string;

  /** ThirdParty's website URL */
  websiteUrl: string;

  /** URL to third party's privacy policy */
  privacyPolicyUrl?: string;

  /** URL to third party's terms of service */
  termsOfServiceUrl?: string;

  /** URL to third party's service level agreement */
  serviceLevelAgreementUrl?: string;

  /** URL to service software agreement */
  serviceSoftwareAgreementUrl?: string;

  /** URL to third party's data processing agreement */
  dataProcessingAgreementUrl?: string;

  /** URL to third party's list of subprocessors */
  subprocessorsListUrl?: string;

  /** URL to third party's business associate agreement */
  businessAssociateAgreementUrl?: string;

  /** Short description of the third party/service */
  description?: string;

  /** Primary category for the third party */
  category?: ThirdPartyCategory;

  /** Security or compliance certifications held by the third party */
  certifications?: string[];

  /** URL to third party's security page */
  securityPageUrl?: string;

  /** URL to third party's trust page */
  trustPageUrl?: string;

  /** URL to third party's status page */
  statusPageUrl?: string;

  /** Countries where the third party is located */
  countries?: CountryCode[];
}

/**
 * Array of third party data
 */
export type ThirdPartys = ThirdParty[];

/**
 * Default export representing the entire third party dataset
 */
declare const data: ThirdPartys;

export default data;
