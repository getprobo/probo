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

import {
  ArrowsClockwiseIcon,
  CheckCircleIcon,
  ClockCountdownIcon,
  HourglassIcon,
  SpinnerGapIcon,
  WarningCircleIcon,
} from "@phosphor-icons/react";
import type { ReactNode } from "react";

export type CustomDomainSslStatus
  = | "ACTIVE"
    | "PROVISIONING"
    | "RENEWING"
    | "PENDING"
    | "FAILED"
    | "EXPIRED";

export type CustomDomainCardTone = "green" | "sky" | "amber" | "red";

export function customDomainCardTone(sslStatus: CustomDomainSslStatus): CustomDomainCardTone {
  switch (sslStatus) {
    case "ACTIVE":
      return "green";
    case "RENEWING":
    case "PROVISIONING":
      return "sky";
    case "PENDING":
      return "amber";
    case "FAILED":
    case "EXPIRED":
      return "red";
  }
}

export function showsCustomDomainDns(managed: boolean): boolean {
  return !managed;
}

export function CustomDomainStatusIcon({ status }: { status: CustomDomainSslStatus }): ReactNode {
  switch (status) {
    case "ACTIVE":
      return <CheckCircleIcon size={32} weight="duotone" />;
    case "RENEWING":
      return <ArrowsClockwiseIcon size={32} weight="duotone" />;
    case "PROVISIONING":
      return <SpinnerGapIcon size={32} weight="duotone" />;
    case "PENDING":
      return <HourglassIcon size={32} weight="duotone" />;
    case "FAILED":
      return <WarningCircleIcon size={32} weight="duotone" />;
    case "EXPIRED":
      return <ClockCountdownIcon size={32} weight="duotone" />;
  }
}
