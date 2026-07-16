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

import type { Icon, IconProps } from "@phosphor-icons/react";
import {
  BellIcon,
  BugIcon,
  CertificateIcon,
  CloudIcon,
  CodeIcon,
  DatabaseIcon,
  EyeIcon,
  EyeSlashIcon,
  FingerprintIcon,
  GavelIcon,
  GlobeIcon,
  HardDrivesIcon,
  HeartbeatIcon,
  KeyIcon,
  LockIcon,
  LockKeyIcon,
  ShieldCheckIcon,
  ShieldWarningIcon,
  SirenIcon,
  UsersIcon,
} from "@phosphor-icons/react";
import { createElement } from "react";

// Maps the CompliancePortalCommitmentIcon enum (coredata) to the Phosphor icon
// rendered on each commitment card. Keep in sync with the console icon picker
// (see apps/console .../commitments/_lib/commitmentIcons.ts).
const COMMITMENT_ICONS: Record<string, Icon> = {
  LOCK_KEY: LockKeyIcon,
  EYE_SLASH: EyeSlashIcon,
  FINGERPRINT: FingerprintIcon,
  SHIELD_WARNING: ShieldWarningIcon,
  SHIELD_CHECK: ShieldCheckIcon,
  SIREN: SirenIcon,
  KEY: KeyIcon,
  LOCK: LockIcon,
  CLOUD: CloudIcon,
  DATABASE: DatabaseIcon,
  GLOBE: GlobeIcon,
  EYE: EyeIcon,
  USERS: UsersIcon,
  CERTIFICATE: CertificateIcon,
  GAVEL: GavelIcon,
  HEARTBEAT: HeartbeatIcon,
  BELL: BellIcon,
  BUG: BugIcon,
  CODE: CodeIcon,
  SERVER: HardDrivesIcon,
};

// Renders the Phosphor icon mapped from a CompliancePortalCommitmentIcon enum
// value. Declared at module scope (and built via createElement) so the icon
// component is never created during a parent's render.
export function CommitmentIcon({ icon, ...props }: { icon: string } & IconProps) {
  const resolved: Icon = COMMITMENT_ICONS[icon] ?? LockKeyIcon;
  return createElement(resolved, props);
}
