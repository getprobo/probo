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
      {logoDownloadUrl && (
        <img
          src={logoDownloadUrl}
          alt=""
          className="h-12 w-auto max-w-[180px] object-contain"
        />
      )}

      <p className="text-lg font-semibold">{name}</p>

      {clientURL && (
        <a
          href={clientURL}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm text-txt-tertiary hover:text-txt-primary transition-colors"
        >
          {clientURLHost(clientURL)}
        </a>
      )}
    </div>
  );
}
