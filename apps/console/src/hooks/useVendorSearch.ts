import type { Vendor } from "@probo/vendors";
import MiniSearch from "minisearch";
import { useEffect, useRef, useState } from "react";

export function useVendorSearch() {
  const searchRef = useRef<(s: string) => Vendor[]>(() => []);
  const [query, setQuery] = useState("");
  const [vendors, setVendors] = useState<Vendor[]>([]);
  useEffect(() => {
    void import("@probo/vendors").then((module) => {
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
    vendors: vendors.slice(0, 20),
    search: (s: string) => {
      setQuery(s);
      setVendors(searchRef.current(s));
    },
  };
}
