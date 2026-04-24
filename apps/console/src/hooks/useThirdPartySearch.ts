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

import type { ThirdParty } from "@probo/third-parties";
import MiniSearch from "minisearch";
import { useEffect, useRef, useState } from "react";

export function useThirdPartySearch() {
  const searchRef = useRef<(s: string) => ThirdParty[]>(() => []);
  const [query, setQuery] = useState("");
  const [thirdParties, setThirdParties] = useState<ThirdParty[]>([]);
  useEffect(() => {
    void import("@probo/third-parties").then((module) => {
      const ms = new MiniSearch({
        fields: ["name"],
        storeFields: Object.keys(module.default[0]),
        searchOptions: {
          fuzzy: 0.1,
          prefix: true,
        },
      });
      ms.addAll(module.default.map(v => ({ ...v, id: v.name })));
      // @ts-expect-error not enough types to handle this case
      searchRef.current = ms.search.bind(ms);
    });
  }, []);

  return {
    query,
    thirdParties: thirdParties.slice(0, 20),
    search: (s: string) => {
      setQuery(s);
      setThirdParties(searchRef.current(s));
    },
  };
}
