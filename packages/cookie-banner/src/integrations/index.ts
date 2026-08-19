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

export type { ConsentIntegration } from "./integration";
export { GoogleConsentModeIntegration } from "./gcm";

import type { IntegrationConfig } from "../types";
import type { ConsentIntegration } from "./integration";
import { GoogleConsentModeIntegration } from "./gcm";

export function createDefaultIntegrations(
  configs?: IntegrationConfig[],
): ConsentIntegration[] {
  const gcm = configs?.find((config) => config.name === "gcm");
  if (gcm?.enabled === false) {
    return [];
  }

  return [
    new GoogleConsentModeIntegration(),
  ];
}

export function resolveGcmEnabled(value: string | null): boolean {
  if (value == null) {
    return true;
  }

  const normalized = value.trim().toLowerCase();
  if (normalized === "true") {
    return true;
  }
  if (normalized === "false") {
    return false;
  }

  console.warn(
    `[probo] invalid gcm-enabled value "${value}": expected "true" or "false", falling back to enabled`,
  );
  return true;
}
