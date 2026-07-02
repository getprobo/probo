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
