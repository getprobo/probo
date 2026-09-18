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

import type { BannerConfig, ConsentAction } from "../types";

export interface TCFChoices {
  purposeConsents: number[];
  purposeLegitimateInterests: number[];
  vendorConsents: number[];
  vendorLegitimateInterests: number[];
  specialFeatureOptins: number[];
}

export interface TCFRuntime {
  onConfig(config: BannerConfig, existingTc?: string): void;
  onConsent(action: ConsentAction, config: BannerConfig): string | undefined;
  onUIVisible?(visible: boolean): void;
  setPendingChoices?(choices: TCFChoices): void;
  getPendingChoices?(): TCFChoices | undefined;
  decodeChoices?(tc: string): TCFChoices | null;
}

let runtime: TCFRuntime | null = null;

export function setTCFRuntime(next: TCFRuntime): void {
  runtime = next;
}

export function getTCFRuntime(): TCFRuntime | null {
  return runtime;
}
