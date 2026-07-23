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

type Status =
  | "ACTIVE"
  | "PROVISIONING"
  | "RENEWING"
  | "PENDING"
  | "FAILED"
  | "EXPIRED";

export const getCustomDomainStatusBadgeVariant = (sslStatus: Status) => {
  switch (sslStatus) {
    case "ACTIVE":
      return "success" as const;
    case "PROVISIONING":
    case "RENEWING":
    case "PENDING":
      return "warning" as const;
    case "FAILED":
    case "EXPIRED":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
};

export const getCustomDomainStatusBadgeLabel = (
  sslStatus: Status,
  t: (key: string) => string,
) => {
  if (sslStatus === "ACTIVE") {
    return t("helpers.customDomain.status.active");
  }
  if (sslStatus === "PROVISIONING" || sslStatus === "RENEWING") {
    return t("helpers.customDomain.status.provisioning");
  }
  if (sslStatus === "PENDING") {
    return t("helpers.customDomain.status.pending");
  }
  if (sslStatus === "FAILED") {
    return t("helpers.customDomain.status.failed");
  }
  if (sslStatus === "EXPIRED") {
    return t("helpers.customDomain.status.expired");
  }
  return t("helpers.common.unknown");
};

const provisioningErrorMessages: Record<string, string> = {
  DNS_CNAME: "helpers.customDomain.provisioningError.dnsCname",
  DNS_CAA: "helpers.customDomain.provisioningError.dnsCaa",
  ACME_RATE_LIMITED: "helpers.customDomain.provisioningError.rateLimited",
  ACME_INVALID_ORDER: "helpers.customDomain.provisioningError.invalidOrder",
  ACME_TEMPORARY: "helpers.customDomain.provisioningError.temporary",
  ACME_FAILED: "helpers.customDomain.provisioningError.failed",
};

export const getCertificateProvisioningErrorMessage = (
  provisioningError: string | null | undefined,
  t: (key: string) => string,
) => {
  if (!provisioningError) {
    return null;
  }

  const knownMessage = provisioningErrorMessages[provisioningError];
  if (knownMessage) {
    return t(knownMessage);
  }

  return t("helpers.customDomain.provisioningError.failed");
};
