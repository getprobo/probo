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

import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

export const auditLogActorTypes = ["USER", "API_KEY", "SYSTEM"] as const;
export type AuditLogActorType = (typeof auditLogActorTypes)[number];

export const auditLogOutcomes = ["ALLOWED", "DENIED"] as const;
export type AuditLogOutcome = (typeof auditLogOutcomes)[number];

const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;
const DAY_MS = 24 * 60 * 60 * 1000;

export interface AuditLogListFilter {
  action?: string | null;
  actionPrefix?: string | null;
  actorId?: string | null;
  actorType?: AuditLogActorType | null;
  resourceType?: string | null;
  resourceId?: string | null;
  outcome?: AuditLogOutcome | null;
  requestId?: string | null;
  createdFrom?: string | null;
  createdTo?: string | null;
}

function isAuditLogActorType(value: string): value is AuditLogActorType {
  return (auditLogActorTypes as readonly string[]).includes(value);
}

function isAuditLogOutcome(value: string): value is AuditLogOutcome {
  return (auditLogOutcomes as readonly string[]).includes(value);
}

function parseEnum<T extends string>(
  value: string,
  isValue: (value: string) => value is T,
): T | null {
  const upper = value.toUpperCase();
  return isValue(upper) ? upper : null;
}

export function isIsoDate(value: string): boolean {
  if (!ISO_DATE.test(value)) {
    return false;
  }
  return !Number.isNaN(Date.parse(`${value}T00:00:00Z`));
}

function utcStart(date: string): string | null {
  if (!isIsoDate(date)) {
    return null;
  }
  return new Date(`${date}T00:00:00Z`).toISOString();
}

function utcEndExclusive(date: string): string | null {
  const start = utcStart(date);
  if (start == null) {
    return null;
  }
  return new Date(Date.parse(start) + DAY_MS).toISOString();
}

function auditLogGraphqlFilter(filters: {
  action: string;
  actorType: AuditLogActorType | null;
  outcome: AuditLogOutcome | null;
  actorId: string;
  resourceType: string;
  resourceId: string;
  from: string;
  to: string;
  requestId: string;
}): AuditLogListFilter | null {
  const filter: AuditLogListFilter = {};

  if (filters.action.endsWith("*")) {
    const prefix = filters.action.slice(0, -1);
    if (prefix !== "") {
      filter.actionPrefix = prefix;
    }
  } else if (filters.action !== "") {
    filter.action = filters.action;
  }

  if (filters.actorType != null) {
    filter.actorType = filters.actorType;
  }
  if (filters.outcome != null) {
    filter.outcome = filters.outcome;
  }
  if (filters.actorId !== "") {
    filter.actorId = filters.actorId;
  }
  if (filters.resourceType !== "") {
    filter.resourceType = filters.resourceType;
  }
  if (filters.resourceId !== "") {
    filter.resourceId = filters.resourceId;
  }
  if (filters.requestId !== "") {
    filter.requestId = filters.requestId;
  }

  const span = filters.from !== "" && filters.to !== "" && filters.from > filters.to
    ? { from: filters.to, to: filters.from }
    : { from: filters.from, to: filters.to };
  const createdFrom = utcStart(span.from);
  const createdTo = utcEndExclusive(span.to);
  if (createdFrom != null) {
    filter.createdFrom = createdFrom;
  }
  if (createdTo != null) {
    filter.createdTo = createdTo;
  }

  return Object.keys(filter).length === 0 ? null : filter;
}

function assignParam(
  previous: URLSearchParams,
  key: string,
  value: string,
): URLSearchParams {
  const next = new URLSearchParams(previous);
  if (value === "") {
    next.delete(key);
  } else {
    next.set(key, value);
  }
  return next;
}

export interface AuditLogListFilters {
  action: string;
  actorType: AuditLogActorType | null;
  outcome: AuditLogOutcome | null;
  actorId: string;
  resourceType: string;
  resourceId: string;
  from: string;
  to: string;
  requestId: string;
  graphqlFilter: AuditLogListFilter | null;
  hasActiveFilters: boolean;
  setAction: (value: string) => void;
  setActorType: (value: AuditLogActorType | null) => void;
  setOutcome: (value: AuditLogOutcome | null) => void;
  setActorId: (value: string) => void;
  setResourceType: (value: string) => void;
  setResourceId: (value: string) => void;
  setFrom: (value: string) => void;
  setTo: (value: string) => void;
  setDates: (from: string, to: string) => void;
  setRequestId: (value: string) => void;
}

export function useAuditLogListFilters(): AuditLogListFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const action = searchParams.get("action") ?? "";
  const actorType = parseEnum(searchParams.get("actor") ?? "", isAuditLogActorType);
  const outcome = parseEnum(searchParams.get("outcome") ?? "", isAuditLogOutcome);
  const actorId = searchParams.get("actorId") ?? "";
  const resourceType = searchParams.get("resourceType") ?? "";
  const resourceId = searchParams.get("resourceId") ?? "";
  const rawFrom = searchParams.get("from") ?? "";
  const rawTo = searchParams.get("to") ?? "";
  const from = isIsoDate(rawFrom) ? rawFrom : "";
  const to = isIsoDate(rawTo) ? rawTo : "";
  const requestId = searchParams.get("request") ?? "";

  const graphqlFilter = useMemo(
    () => auditLogGraphqlFilter({
      action,
      actorType,
      outcome,
      actorId,
      resourceType,
      resourceId,
      from,
      to,
      requestId,
    }),
    [action, actorType, outcome, actorId, resourceType, resourceId, from, to, requestId],
  );

  const setParam = useCallback((key: string, value: string) => {
    setSearchParams(previous => assignParam(previous, key, value), { replace: true });
  }, [setSearchParams]);

  const setAction = useCallback((value: string) => {
    setParam("action", value);
  }, [setParam]);

  const setActorType = useCallback((value: AuditLogActorType | null) => {
    setParam("actor", value == null ? "" : value.toLowerCase());
  }, [setParam]);

  const setOutcome = useCallback((value: AuditLogOutcome | null) => {
    setParam("outcome", value == null ? "" : value.toLowerCase());
  }, [setParam]);

  const setActorId = useCallback((value: string) => {
    setParam("actorId", value);
  }, [setParam]);

  const setResourceType = useCallback((value: string) => {
    setParam("resourceType", value);
  }, [setParam]);

  const setResourceId = useCallback((value: string) => {
    setParam("resourceId", value);
  }, [setParam]);

  const setFrom = useCallback((value: string) => {
    setParam("from", value);
  }, [setParam]);

  const setTo = useCallback((value: string) => {
    setParam("to", value);
  }, [setParam]);

  const setDates = useCallback((from: string, to: string) => {
    setSearchParams((previous) => {
      const withFrom = assignParam(previous, "from", from);
      return assignParam(withFrom, "to", to);
    }, { replace: true });
  }, [setSearchParams]);

  const setRequestId = useCallback((value: string) => {
    setParam("request", value);
  }, [setParam]);

  return {
    action,
    actorType,
    outcome,
    actorId,
    resourceType,
    resourceId,
    from,
    to,
    requestId,
    graphqlFilter,
    hasActiveFilters: graphqlFilter != null,
    setAction,
    setActorType,
    setOutcome,
    setActorId,
    setResourceType,
    setResourceId,
    setFrom,
    setTo,
    setDates,
    setRequestId,
  };
}
