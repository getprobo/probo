// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { Text } from "@probo/ui/src/v2/typography/Text";

type OAuthClientBrandingSectionProps = {
  name: string;
  logoDownloadUrl?: string | null;
  clientURL?: string | null;
};

function clientURLHost(clientURL: string): string {
  try {
    return new URL(clientURL).host;
  } catch {
    return clientURL;
  }
}

export function OAuthClientBrandingSection({
  name,
  logoDownloadUrl,
  clientURL,
}: OAuthClientBrandingSectionProps) {
  return (
    <div className="flex flex-col items-center gap-3 text-center">
      {logoDownloadUrl
        ? (
            <img
              src={logoDownloadUrl}
              alt=""
              className="h-12 w-auto max-w-45 object-contain"
            />
          )
        : (
            <Avatar size={4} fallback={name.slice(0, 1)} />
          )}

      <Text size={4} weight="medium" highContrast className="block">
        {name}
      </Text>

      {clientURL && (
        <Anchor
          href={clientURL}
          target="_blank"
          rel="noopener noreferrer"
          size={2}
        >
          {clientURLHost(clientURL)}
        </Anchor>
      )}
    </div>
  );
}
