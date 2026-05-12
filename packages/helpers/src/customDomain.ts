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
  __: (key: string) => string,
) => {
  if (sslStatus === "ACTIVE") {
    return __("Active");
  }
  if (sslStatus === "PROVISIONING" || sslStatus === "RENEWING") {
    return __("Provisioning");
  }
  if (sslStatus === "PENDING") {
    return __("Pending");
  }
  if (sslStatus === "FAILED") {
    return __("Failed");
  }
  if (sslStatus === "EXPIRED") {
    return __("Expired");
  }
  return __("Unknown");
};
