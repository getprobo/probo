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

import type { CompliancePortalCommitmentIcon } from "#/__generated__/core/CompliancePageCommitmentDialogCreateMutation.graphql";

// The curated icon set mirrors coredata.CompliancePortalCommitmentIcon and the
// Phosphor icons the compliance portal renders for each value.
export const COMMITMENT_ICON_VALUES: CompliancePortalCommitmentIcon[] = [
  "LOCK_KEY",
  "EYE_SLASH",
  "FINGERPRINT",
  "SHIELD_WARNING",
  "SHIELD_CHECK",
  "SIREN",
  "KEY",
  "LOCK",
  "CLOUD",
  "DATABASE",
  "GLOBE",
  "EYE",
  "USERS",
  "CERTIFICATE",
  "GAVEL",
  "HEARTBEAT",
  "BELL",
  "BUG",
  "CODE",
  "SERVER",
];

// Human-readable label for each icon value, shown in the console picker.
export const COMMITMENT_ICON_LABELS: Record<CompliancePortalCommitmentIcon, string> = {
  LOCK_KEY: "Lock & key",
  EYE_SLASH: "Eye slash",
  FINGERPRINT: "Fingerprint",
  SHIELD_WARNING: "Shield warning",
  SHIELD_CHECK: "Shield check",
  SIREN: "Siren",
  KEY: "Key",
  LOCK: "Lock",
  CLOUD: "Cloud",
  DATABASE: "Database",
  GLOBE: "Globe",
  EYE: "Eye",
  USERS: "Users",
  CERTIFICATE: "Certificate",
  GAVEL: "Gavel",
  HEARTBEAT: "Heartbeat",
  BELL: "Bell",
  BUG: "Bug",
  CODE: "Code",
  SERVER: "Server",
};
