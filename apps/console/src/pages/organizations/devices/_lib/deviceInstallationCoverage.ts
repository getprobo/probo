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

export interface DeviceInstallationDevice {
  state: string;
  ownerId: string | null;
}

export interface DeviceInstallationCoverage {
  total: number;
  covered: number;
  missing: number;
  tone: "green" | "amber";
}

export function deviceInstallationCoverage(
  employeeIds: readonly string[],
  devices: readonly DeviceInstallationDevice[],
): DeviceInstallationCoverage {
  const employees = new Set(employeeIds);
  const coveredOwners = new Set<string>();
  for (const device of devices) {
    if (
      device.state === "ACTIVE"
      && device.ownerId != null
      && employees.has(device.ownerId)
    ) {
      coveredOwners.add(device.ownerId);
    }
  }
  const total = employees.size;
  const covered = coveredOwners.size;
  const missing = total - covered;
  return {
    total,
    covered,
    missing,
    tone: missing === 0 ? "green" : "amber",
  };
}
