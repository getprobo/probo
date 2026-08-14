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

import { lazy } from "@probo/react-lazy";
import type { ComponentType } from "react";

/**
 * Feature switchers, loaded only when the active panel renders that item.
 *
 * Keep this as lazy() loaders, never static imports: the organization layout
 * must not pull cookie-banner (or later) Relay artifacts into the shell chunk.
 */
export const navPanelSwitchers = {
  "cookie-banners": lazy(async () => {
    const { CookieBannerSwitcher } = await import(
      "#/pages/organizations/cookie-banners/_components/CookieBannerSwitcher"
    );
    return { default: CookieBannerSwitcher as ComponentType };
  }),
  "compliance-portals": lazy(async () => {
    const { CompliancePortalSwitcher } = await import(
      "#/pages/organizations/compliance-portals/_components/CompliancePortalSwitcher"
    );
    return { default: CompliancePortalSwitcher as ComponentType };
  }),
} as const;

export function navPanelSwitcher(path: string) {
  if (!(path in navPanelSwitchers)) {
    throw new Error(`missing nav panel switcher for ${path}`);
  }
  return navPanelSwitchers[path as keyof typeof navPanelSwitchers];
}
