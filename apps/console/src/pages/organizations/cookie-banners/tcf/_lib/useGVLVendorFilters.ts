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

import { useCallback } from "react";
import { useSearchParams } from "react-router";

export const gvlVendorMembershipOptions = ["on", "off"] as const;

export type GVLVendorMembershipOption = (typeof gvlVendorMembershipOptions)[number];
export type GVLVendorMembership = "all" | GVLVendorMembershipOption;

const graphqlMemberships = {
  on: "ON_BANNER",
  off: "NOT_ON_BANNER",
} as const;

export function isGVLVendorMembershipOption(value: string): value is GVLVendorMembershipOption {
  return (gvlVendorMembershipOptions as readonly string[]).includes(value);
}

export function gvlVendorGraphqlFilter(
  query: string,
  membership: GVLVendorMembership,
  cookieBannerId: string,
) {
  const trimmedQuery = query || null;
  const graphqlMembership = membership === "all" ? null : graphqlMemberships[membership];
  if (trimmedQuery == null && graphqlMembership == null) {
    return null;
  }

  return {
    query: trimmedQuery,
    cookieBannerId: graphqlMembership == null ? null : cookieBannerId,
    membership: graphqlMembership,
  };
}

export interface GVLVendorFilters {
  query: string;
  membership: GVLVendorMembership;
  hasActiveFilters: boolean;
  setQuery: (value: string) => void;
  setMembership: (value: GVLVendorMembership) => void;
  clear: () => void;
}

export function useGVLVendorFilters(): GVLVendorFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const rawMembership = searchParams.get("membership") ?? "";
  const membership: GVLVendorMembership = isGVLVendorMembershipOption(rawMembership)
    ? rawMembership
    : "all";

  const setParam = useCallback((key: string, value: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set(key, value);
      } else {
        next.delete(key);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setQuery = useCallback((value: string) => setParam("q", value), [setParam]);

  const setMembership = useCallback((value: GVLVendorMembership) => {
    setParam("membership", value === "all" ? "" : value);
  }, [setParam]);

  const clear = useCallback(() => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      next.delete("q");
      next.delete("membership");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    query,
    membership,
    hasActiveFilters: query !== "" || membership !== "all",
    setQuery,
    setMembership,
    clear,
  };
}
