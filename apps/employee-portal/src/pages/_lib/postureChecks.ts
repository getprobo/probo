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

type Translator = (key: string) => string;

const checkKeyLabels: Record<string, string> = {
  DISK_ENCRYPTION: "homePage.dashboard.devices.checks.diskEncryption",
  SCREEN_LOCK: "homePage.dashboard.devices.checks.screenLock",
  FIREWALL_ENABLED: "homePage.dashboard.devices.checks.firewallEnabled",
  TIME_SYNC: "homePage.dashboard.devices.checks.timeSync",
  OS_VERSION: "homePage.dashboard.devices.checks.osVersion",
  AUTO_UPDATE: "homePage.dashboard.devices.checks.autoUpdate",
  PASSWORD_POLICY: "homePage.dashboard.devices.checks.passwordPolicy",
  REMOTE_LOGIN: "homePage.dashboard.devices.checks.remoteLogin",
  MALWARE_PROTECTION: "homePage.dashboard.devices.checks.malwareProtection",
};

export type PostureOutcome = "pass" | "warning" | "unknown";

// Stored status is the verdict. NOT_APPLICABLE is not a warning.
// A status this client does not know stays visible as unknown.
export function postureStatusOutcome(
  status: string,
): PostureOutcome | "notApplicable" {
  switch (status) {
    case "PASS":
      return "pass";
    case "FAIL":
      return "warning";
    case "UNKNOWN":
      return "unknown";
    case "NOT_APPLICABLE":
      return "notApplicable";
    default:
      return "unknown";
  }
}

export function getPostureCheckLabel(t: Translator, checkKey: string) {
  const label = checkKeyLabels[checkKey];
  return label ? t(label) : checkKey;
}
