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

import { graphql } from "react-relay";
import { type LiveState, readFragment } from "relay-runtime";

import type { TrustCenterLogoResolverFragment$key } from "./__generated__/TrustCenterLogoResolverFragment.graphql";

function prefersDark(): boolean {
  return typeof window !== "undefined"
    && !!window.matchMedia
    && window.matchMedia("(prefers-color-scheme: dark)").matches;
}

/**
 * @relayField TrustCenter.themedLogoUrl: String
 * @rootFragment TrustCenterLogoResolverFragment
 * @live
 *
 * Resolves the trust center logo download URL for the current system color
 * scheme: the dark logo (falling back to the light one) when the OS prefers
 * dark, otherwise the light logo. Lives in the graph so consumers select a
 * single field instead of threading `useSystemTheme` through URL selection.
 */
export function themedLogoUrl(
  key: TrustCenterLogoResolverFragment$key,
): LiveState<string | null> {
  const data = readFragment(
    graphql`
      fragment TrustCenterLogoResolverFragment on TrustCenter {
        logo {
          downloadUrl
        }
        darkLogo {
          downloadUrl
        }
      }
    `,
    key,
  );
  const lightUrl = data.logo?.downloadUrl ?? null;
  const darkUrl = data.darkLogo?.downloadUrl ?? lightUrl;

  return {
    read: () => (prefersDark() ? darkUrl : lightUrl),
    subscribe: (callback) => {
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
      mediaQuery.addEventListener("change", callback);
      return () => mediaQuery.removeEventListener("change", callback);
    },
  };
}
