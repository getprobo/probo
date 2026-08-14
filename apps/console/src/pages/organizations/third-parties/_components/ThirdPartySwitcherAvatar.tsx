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

import { faviconUrl } from "@probo/helpers";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";

export interface ThirdPartySwitcherAvatarProps {
  name: string;
  websiteUrl: string | null | undefined;
}

/**
 * Favicon (or initial) for a third party in the TPRM-panel switcher.
 *
 * Same source as the list row: Google's favicon service keyed off the site
 * hostname, falling back to the first letter when there is no URL.
 */
export function ThirdPartySwitcherAvatar({ name, websiteUrl }: ThirdPartySwitcherAvatarProps) {
  return (
    <Avatar
      size={1}
      variant="soft"
      color="neutral"
      radius="small"
      src={faviconUrl(websiteUrl) ?? undefined}
      alt={name}
      fallback={name.charAt(0) || "?"}
    />
  );
}
