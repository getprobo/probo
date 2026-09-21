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

import type { BannerConfig, BannerState } from "../types";

export interface LayoutHost {
  readonly bannerConfig: BannerConfig;
  readonly consentDraft: Record<string, boolean>;
  setState(state: BannerState): void;
  dispatchEvent(event: Event): boolean;
  readonly client: {
    customize(categories: Record<string, boolean>): void;
    acceptAll?(): void;
    rejectAll?(): void;
  };
}

export interface LayoutRenderer {
  render(config: BannerConfig, position: string): string | null;
  wire?(root: LayoutHost, host: ShadowRoot): void;
}

let renderer: LayoutRenderer | null = null;

export function setLayoutRenderer(next: LayoutRenderer | null): void {
  renderer = next;
}

export function getLayoutRenderer(): LayoutRenderer | null {
  return renderer;
}
