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

import { TCString } from "@iabtechlabtcf/core";

export interface DecodedTcString {
  created: Date;
  lastUpdated: Date;
  cmpId: string;
  cmpVersion: string;
  tcfPolicyVersion: string;
  gvlVersion: string;
  language: string;
  purposeIds: number[];
  vendorConsentCount: number;
}

function vectorIds(vector: { maxId: number; has(id: number): boolean }): number[] {
  const ids: number[] = [];
  for (let id = 1; id <= vector.maxId; id++) {
    if (vector.has(id)) {
      ids.push(id);
    }
  }
  return ids;
}

export function decodeTcString(encoded: string): DecodedTcString | null {
  try {
    const model = TCString.decode(encoded);
    const purposeIds = vectorIds(model.purposeConsents);
    return {
      created: model.created,
      lastUpdated: model.lastUpdated,
      cmpId: String(model.cmpId),
      cmpVersion: String(model.cmpVersion),
      tcfPolicyVersion: String(model.policyVersion),
      gvlVersion: model.vendorListVersion === 0 ? "-" : String(model.vendorListVersion),
      language: model.consentLanguage,
      purposeIds,
      vendorConsentCount: vectorIds(model.vendorConsents).length,
    };
  } catch {
    return null;
  }
}
