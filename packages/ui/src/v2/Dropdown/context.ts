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

import { createContext, useContext } from "react";

// Menu-level look (set on the popup) shared with every item, so item styling
// stays consistent without threading props through each one.
export type DropdownContextValue = {
  size: 1 | 2;
  variant: "solid" | "soft";
  highContrast: boolean;
};

const DropdownContext = createContext<DropdownContextValue>({
  size: 2,
  variant: "solid",
  highContrast: false,
});

export const DropdownProvider = DropdownContext.Provider;

export function useDropdownContext() {
  return useContext(DropdownContext);
}
