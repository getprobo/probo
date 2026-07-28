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

type Translator = (key: string) => string;

const checkKeyLabels: Record<string, string> = {
  DISK_ENCRYPTION: "devices.postures.checks.diskEncryption",
  SCREEN_LOCK: "devices.postures.checks.screenLock",
  FIREWALL_ENABLED: "devices.postures.checks.firewallEnabled",
  TIME_SYNC: "devices.postures.checks.timeSync",
  OS_VERSION: "devices.postures.checks.osVersion",
  AUTO_UPDATE: "devices.postures.checks.autoUpdate",
  PASSWORD_POLICY: "devices.postures.checks.passwordPolicy",
  REMOTE_LOGIN: "devices.postures.checks.remoteLogin",
  MALWARE_PROTECTION: "devices.postures.checks.malwareProtection",
};

export function getPostureCheckLabel(t: Translator, checkKey: string) {
  const label = checkKeyLabels[checkKey];
  return label ? t(label) : checkKey;
}
