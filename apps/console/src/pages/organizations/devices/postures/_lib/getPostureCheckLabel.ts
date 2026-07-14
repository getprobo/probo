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

type Translator = (s: string) => string;

const checkKeyLabels: Record<string, string> = {
  DISK_ENCRYPTION: "Disk encryption",
  SCREEN_LOCK: "Screen lock",
  FIREWALL_ENABLED: "Firewall enabled",
  TIME_SYNC: "Time sync",
  OS_VERSION: "OS version",
  AUTO_UPDATE: "Auto update",
  PASSWORD_POLICY: "Password policy",
  REMOTE_LOGIN: "Remote login",
  MALWARE_PROTECTION: "Malware protection",
};

export function getPostureCheckLabel(__: Translator, checkKey: string) {
  const label = checkKeyLabels[checkKey];
  return label ? __(label) : checkKey;
}
