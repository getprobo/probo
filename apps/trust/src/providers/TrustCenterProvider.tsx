// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { createContext, type ReactNode } from "react";

import type { TrustGraphCurrentQuery$data } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";

export const TrustCenterContext = createContext<
  TrustGraphCurrentQuery$data["currentTrustCenter"] | null
>(null);

export const TrustCenterProvider = ({
  children,
  trustCenter,
}: {
  children: ReactNode;
  trustCenter: TrustGraphCurrentQuery$data["currentTrustCenter"];
}) => {
  return (
    <TrustCenterContext.Provider value={trustCenter}>
      {children}
    </TrustCenterContext.Provider>
  );
};
