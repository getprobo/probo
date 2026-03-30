// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { IconBrandFacebook } from "./IconBrandFacebook";
import { IconBrandLinkedin } from "./IconBrandLinkedin";
import { IconBrandX } from "./IconBrandX";
import { IconGlobe } from "./IconGlobe";
import type { IconProps } from "./type";

export type SocialIconProps = IconProps & { colored?: boolean };

export function SocialIcon({ socialName, ...props }: { socialName: string | null } & SocialIconProps) {
  switch (socialName) {
    case "LinkedIn": return <IconBrandLinkedin {...props} />;
    case "X": return <IconBrandX {...props} />;
    case "Facebook": return <IconBrandFacebook {...props} />;
    default: return <IconGlobe {...props} />;
  }
}
