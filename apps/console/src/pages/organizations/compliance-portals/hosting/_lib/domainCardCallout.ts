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

import type { CustomDomainSslStatus } from "./customDomainCardTone";

export interface DomainCardCallout {
  title: string;
  description: string;
  detail: string | null;
  dns: string | null;
}

function domainCardDnsInstruction(
  sslStatus: CustomDomainSslStatus,
  t: (key: string) => string,
): string | null {
  switch (sslStatus) {
    case "PENDING":
      return t("domainCard.pending.dns");
    case "PROVISIONING":
      return t("domainCard.provisioning.dns");
    case "FAILED":
      return t("domainCard.failed.dns");
    case "EXPIRED":
      return t("domainCard.expired.dns");
    case "ACTIVE":
      return t("domainCard.active.dns");
    case "RENEWING":
      return t("domainCard.renewing.dns");
  }
}

export function domainCardCallout(
  sslStatus: CustomDomainSslStatus,
  provisioningErrorMessage: string | null,
  expiresLabel: string | null,
  t: (key: string, options?: Record<string, string>) => string,
): DomainCardCallout {
  const dns = domainCardDnsInstruction(sslStatus, t);

  if (provisioningErrorMessage != null) {
    return {
      title: t("domainCard.error.title"),
      description: provisioningErrorMessage,
      detail: null,
      dns,
    };
  }

  switch (sslStatus) {
    case "ACTIVE":
      return {
        title: t("domainCard.active.title"),
        description: t("domainCard.active.description"),
        detail: expiresLabel == null
          ? null
          : t("domainCard.sslExpires", { date: expiresLabel }),
        dns,
      };
    case "PROVISIONING":
      return {
        title: t("domainCard.provisioning.title"),
        description: t("domainCard.provisioning.description"),
        detail: null,
        dns,
      };
    case "RENEWING":
      return {
        title: t("domainCard.renewing.title"),
        description: t("domainCard.renewing.description"),
        detail: null,
        dns,
      };
    case "EXPIRED":
      return {
        title: t("domainCard.expired.title"),
        description: t("domainCard.expired.description"),
        detail: null,
        dns,
      };
    case "FAILED":
      return {
        title: t("domainCard.failed.title"),
        description: t("domainCard.failed.description"),
        detail: null,
        dns,
      };
    case "PENDING":
      return {
        title: t("domainCard.pending.title"),
        description: t("domainCard.pending.description"),
        detail: null,
        dns,
      };
  }
}
