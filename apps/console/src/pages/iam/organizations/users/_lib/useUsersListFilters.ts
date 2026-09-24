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

import { peopleRoles, Role, roles } from "@probo/helpers";
import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

export const usersListProfileStates = ["PENDING", "ACTIVE", "DEACTIVATED"] as const;
export type UsersListProfileState = (typeof usersListProfileStates)[number];

const usersListStatusOptions = ["pending", "active", "deactivated"] as const;
type UsersListStatusOption = (typeof usersListStatusOptions)[number];
export type UsersListKind = (typeof peopleRoles)[number];

const graphqlStates = {
  pending: "PENDING",
  active: "ACTIVE",
  deactivated: "DEACTIVATED",
} as const;

function isUsersListStatusOption(value: string): value is UsersListStatusOption {
  return (usersListStatusOptions as readonly string[]).includes(value);
}

export function isUsersListRole(value: string): value is Role {
  return (roles as readonly string[]).includes(value);
}

export function isUsersListKind(value: string): value is UsersListKind {
  return (peopleRoles as readonly string[]).includes(value);
}

function usersListGraphqlFilter(filters: {
  query: string;
  status: UsersListProfileState | null;
  role: Role | null;
  kind: UsersListKind | null;
}) {
  return {
    query: filters.query === "" ? null : filters.query,
    states: filters.status == null ? null : [filters.status],
    role: filters.role,
    kind: filters.kind,
    contractEnded: null,
  };
}

export interface UsersListFilters {
  query: string;
  status: UsersListProfileState | null;
  role: Role | null;
  kind: UsersListKind | null;
  graphqlFilter: ReturnType<typeof usersListGraphqlFilter>;
  hasActiveFilters: boolean;
  setQuery: (value: string) => void;
  setStatus: (value: UsersListProfileState | null) => void;
  setRole: (value: Role | null) => void;
  setKind: (value: UsersListKind | null) => void;
}

export function useUsersListFilters(): UsersListFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const rawStatus = searchParams.get("status") ?? "";
  const rawRole = searchParams.get("role") ?? "";
  const rawKind = searchParams.get("kind") ?? "";
  const status = isUsersListStatusOption(rawStatus) ? graphqlStates[rawStatus] : null;
  const role = isUsersListRole(rawRole) ? rawRole : null;
  const kind = isUsersListKind(rawKind) ? rawKind : null;

  const graphqlFilter = useMemo(
    () => usersListGraphqlFilter({ query, status, role, kind }),
    [query, status, role, kind],
  );

  const setQuery = useCallback((value: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set("q", value);
      } else {
        next.delete("q");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setStatus = useCallback((value: UsersListProfileState | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set("status", value.toLowerCase());
      } else {
        next.delete("status");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setRole = useCallback((value: Role | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set("role", value);
      } else {
        next.delete("role");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setKind = useCallback((value: UsersListKind | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set("kind", value);
      } else {
        next.delete("kind");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    query,
    status,
    role,
    kind,
    graphqlFilter,
    hasActiveFilters: query !== "" || status != null || role != null || kind != null,
    setQuery,
    setStatus,
    setRole,
    setKind,
  };
}
