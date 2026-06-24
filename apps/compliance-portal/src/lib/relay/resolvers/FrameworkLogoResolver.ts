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

import { graphql } from "react-relay";
import { type LiveState, readFragment } from "relay-runtime";

import type { FrameworkLogoResolverFragment$key } from "./__generated__/FrameworkLogoResolverFragment.graphql";

function prefersDark(): boolean {
  return typeof window !== "undefined"
    && !!window.matchMedia
    && window.matchMedia("(prefers-color-scheme: dark)").matches;
}

/**
 * @relayField Framework.themedLogoUrl: String
 * @rootFragment FrameworkLogoResolverFragment
 * @live
 *
 * Resolves the framework logo download URL for the current system color scheme:
 * the dark logo (falling back to the light one) when the OS prefers dark,
 * otherwise the light logo. Lives in the graph so consumers select a single
 * field instead of threading `useSystemTheme` through URL selection.
 */
export function themedLogoUrl(
  key: FrameworkLogoResolverFragment$key,
): LiveState<string | null> {
  const data = readFragment(
    graphql`
      fragment FrameworkLogoResolverFragment on Framework {
        lightLogo {
          downloadUrl
        }
        darkLogo {
          downloadUrl
        }
      }
    `,
    key,
  );
  const lightUrl = data.lightLogo?.downloadUrl ?? null;
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
