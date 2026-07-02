// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { Icon } from "@phosphor-icons/react";
import { EyeSlashIcon, FingerprintIcon, LockKeyIcon, ShieldWarningIcon, SirenIcon } from "@phosphor-icons/react";

export interface SecurityCommitmentItem {
  eyebrow: string;
  title: string;
  description: string;
  Icon: Icon;
}

export interface SecurityCommitmentGroup {
  // Only the first group carries the section eyebrow in the design.
  eyebrow?: string;
  title: string;
  description: string;
  items: SecurityCommitmentItem[];
}

// TODO: This content is placeholder data. There is no backend / DB structure for
// security commitments yet — replace this POJO with relay-sourced data (and move
// the copy into i18n) once the schema exists.
export const SECURITY_COMMITMENT_GROUPS: SecurityCommitmentGroup[] = [
  {
    eyebrow: "Security Commitments",
    title: "Data Protection",
    description:
      "Your data is protected with industry-leading encryption and strict privacy controls, from storage to transit to processing.",
    items: [
      {
        eyebrow: "Data Protection",
        title: "Encrypted at rest. Encrypted in transit. Always.",
        description:
          "Customer data is protected with AES-256 at rest and TLS 1.3 in transit. Keys rotated on a fixed schedule.",
        Icon: LockKeyIcon,
      },
      {
        eyebrow: "Privacy",
        title: "Customer data stays customer data.",
        description:
          "PII is masked in non-production environments. Access requires SSO with hardware-backed keys.",
        Icon: EyeSlashIcon,
      },
      {
        eyebrow: "Identity",
        title: "Phishing-resistant authentication, by default.",
        description:
          "Every employee authenticates via WebAuthn. Production access is just-in-time and recorded.",
        Icon: FingerprintIcon,
      },
    ],
  },
  {
    title: "Operational Security",
    description:
      "We maintain continuous uptime and rapid incident response so you can rely on our platform around the clock.",
    items: [
      {
        eyebrow: "Threat Detection",
        title: "Anomalies do not wait. Neither do we.",
        description:
          "24/7 SIEM monitoring with automated triage. Critical incidents page on-call within 90 seconds.",
        Icon: ShieldWarningIcon,
      },
      {
        eyebrow: "Incident Response",
        title: "Issues are caught fast. And handled.",
        description:
          "Security incidents are investigated within hours. Customers notified within 24 hours of a confirmed breach.",
        Icon: SirenIcon,
      },
    ],
  },
];
