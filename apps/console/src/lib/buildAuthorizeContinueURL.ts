// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

const oauth2AuthorizePathSuffix = "/oauth2/authorize";

function parseContinueUrl(continueParam: string | null): URL | null {
  if (!continueParam) {
    return null;
  }

  try {
    return new URL(continueParam, window.location.origin);
  } catch {
    return null;
  }
}

export function isOAuthAuthorizeContinueUrl(continueParam: string | null): boolean {
  const url = parseContinueUrl(continueParam);
  if (!url) {
    return false;
  }

  return url.pathname.endsWith(oauth2AuthorizePathSuffix);
}

export function clientIdFromContinueUrl(continueParam: string | null): string | null {
  const url = parseContinueUrl(continueParam);
  if (!url) {
    return null;
  }

  return url.searchParams.get("client_id");
}
