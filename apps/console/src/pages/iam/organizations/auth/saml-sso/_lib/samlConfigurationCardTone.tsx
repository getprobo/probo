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

import { CheckCircleIcon, HourglassIcon, LockIcon } from "@phosphor-icons/react";
import type { ReactNode } from "react";

import type { TonedCardTone } from "#/components/TonedCard/variants";

export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";

export function samlConfigurationCardTone(
  domainVerifiedAt: string | null | undefined,
  enforcementPolicy: SAMLEnforcementPolicy,
): TonedCardTone {
  if (domainVerifiedAt == null) {
    return "amber";
  }
  if (enforcementPolicy === "OFF") {
    return "sand";
  }
  return "green";
}

export function showsSamlLoginUrl(
  domainVerifiedAt: string | null | undefined,
  enforcementPolicy: SAMLEnforcementPolicy,
): boolean {
  return domainVerifiedAt != null && enforcementPolicy !== "OFF";
}

export function samlConfigurationStatusKey(
  domainVerifiedAt: string | null | undefined,
  enforcementPolicy: SAMLEnforcementPolicy,
): "pending" | "optional" | "required" | "off" {
  if (domainVerifiedAt == null) {
    return "pending";
  }
  switch (enforcementPolicy) {
    case "REQUIRED":
      return "required";
    case "OFF":
      return "off";
    default:
      return "optional";
  }
}

export function SAMLConfigurationStatusIcon({
  domainVerifiedAt,
  enforcementPolicy,
}: {
  domainVerifiedAt: string | null | undefined;
  enforcementPolicy: SAMLEnforcementPolicy;
}): ReactNode {
  if (domainVerifiedAt == null) {
    return <HourglassIcon size={24} weight="duotone" />;
  }
  if (enforcementPolicy === "OFF") {
    return <LockIcon size={24} weight="duotone" />;
  }
  return <CheckCircleIcon size={24} weight="duotone" />;
}
