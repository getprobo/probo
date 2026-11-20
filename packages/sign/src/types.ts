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

export type SigningConfigMode = "LIVE" | "TEST";

export type AcceptanceStatus =
  | "REQUIRED"
  | "ACCEPTED"
  | "REACCEPTANCE_REQUIRED"
  | "ARCHIVED";

export type SigningConfigKind =
  | "TERMS_OF_SERVICE"
  | "ACCEPTABLE_USE_POLICY"
  | "OTHER";

/** All 20 stable error codes from the data plane API. */
export type ErrorCode =
  | "consent_text_mismatch"
  | "version_not_presentable"
  | "version_mismatch"
  | "invalid_email"
  | "invalid_ip_address"
  | "batch_validation_failed"
  | "missing_authorization"
  | "invalid_key"
  | "identity_jwt_expired"
  | "identity_jwt_invalid"
  | "config_not_active"
  | "origin_not_allowed"
  | "test_key_on_live_config"
  | "config_not_found"
  | "subject_not_found"
  | "acceptance_not_found"
  | "already_revoked"
  | "rate_limited"
  | "document_not_presentable"
  | "server_error";

/** Document content returned by GET /documents/{signingConfigId}/content */
export interface DocumentContent {
  documentVersionId: string;
  major: number;
  minor: number;
  contentHtml: string;
  contentSha256: string;
  pdfUrl: string;
  consentText: string;
  kind: SigningConfigKind;
}

/** Per-document acceptance status returned by the status endpoints. */
export interface AcceptanceStatusInfo {
  subjectId?: string;
  externalUserId: string;
  status: AcceptanceStatus;
  acceptedVersionId?: string;
  acceptedAt?: string;
  currentVersionId: string;
  currentVersionMajor: number;
  currentVersionMinor: number;
}

/** Single acceptance result. */
export interface AcceptanceResult {
  acceptanceId: string;
  subjectId: string;
  signingConfigId: string;
  documentVersionId: string;
  acceptedAt: string;
  mode: SigningConfigMode;
  status: "pending_seal";
}

/** Options for initializing the ProboSign client. */
export interface ProboSignOptions {
  /** Publishable key (pk_live_... or pk_test_...) for this signing config. */
  publishableKey: string;
  /**
   * Short-lived identity JWT minted by your backend via signIdentity().
   * Required for accept(); omit for status/loadDocument only.
   */
  identityToken?: string;
  /** Override the API base URL. Defaults to the current origin. */
  baseUrl?: string;
}

/** Decoded identity JWT payload (client-side, unverified). */
export interface IdentityTokenClaims {
  sub: string;
  email: string;
  name?: string;
  account_id?: string;
  iat: number;
  exp: number;
  aud?: string;
}

/** Event map for client.on() */
export interface SignClientEvents {
  accepted: AcceptanceResult[];
}

export type SignEventHandler<K extends keyof SignClientEvents> = (
  data: SignClientEvents[K],
) => void;
