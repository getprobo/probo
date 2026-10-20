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

import {
  type DevicePlatformName,
  devicePlatforms,
} from "./deviceDisplay";

const deviceStates = ["PENDING", "ACTIVE", "REVOKED"] as const;
export type DeviceStateFilter = (typeof deviceStates)[number];

const deviceListViews = ["all", "warnings"] as const;
export type DeviceListView = (typeof deviceListViews)[number];

export function isDeviceListView(value: string): value is DeviceListView {
  return (deviceListViews as readonly string[]).includes(value);
}

export function devicesListOwnerQuery(query: string): string | null {
  const trimmed = query.trim();
  return trimmed === "" ? null : trimmed;
}

function isDeviceStateFilter(value: string): value is DeviceStateFilter {
  return (deviceStates as readonly string[]).includes(value);
}

function isDevicePlatformName(value: string): value is DevicePlatformName {
  return (devicePlatforms as readonly string[]).includes(value);
}

function selectedValues<T extends string>(
  values: readonly string[],
  isAllowed: (value: string) => value is T,
): T[] {
  const seen = new Set<string>();
  const selected: T[] = [];
  for (const value of values) {
    if (seen.has(value) || !isAllowed(value)) {
      continue;
    }
    seen.add(value);
    selected.push(value);
  }
  return selected;
}

function selectedText(values: readonly string[]): string[] {
  return selectedValues(values, (value): value is string => value.length > 0);
}

export function useDevicesListFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const view: DeviceListView = searchParams.get("view") === "warnings" ? "warnings" : "all";
  const states = useMemo(
    () => selectedValues(searchParams.getAll("state"), isDeviceStateFilter),
    [searchParams],
  );
  const platforms = useMemo(
    () => selectedValues(searchParams.getAll("platform"), isDevicePlatformName),
    [searchParams],
  );
  const agentVersions = useMemo(
    () => selectedText(searchParams.getAll("agentVersion")),
    [searchParams],
  );

  const setParam = useCallback((name: string, value: string | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set(name, value);
      } else {
        next.delete(name);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setListParam = useCallback((name: string, values: readonly string[]) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      next.delete(name);
      for (const value of values) {
        next.append(name, value);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setQuery = useCallback((value: string) => {
    setParam("q", value || null);
  }, [setParam]);

  const setView = useCallback((value: DeviceListView) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value === "warnings") {
        next.set("view", "warnings");
      } else {
        next.delete("view");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setStates = useCallback((values: readonly DeviceStateFilter[]) => {
    setListParam("state", values);
  }, [setListParam]);

  const setPlatforms = useCallback((values: readonly DevicePlatformName[]) => {
    setListParam("platform", values);
  }, [setListParam]);

  const setAgentVersions = useCallback((values: readonly string[]) => {
    setListParam("agentVersion", selectedText(values));
  }, [setListParam]);

  return {
    query,
    ownerQuery: devicesListOwnerQuery(query),
    view,
    hasPostureWarning: view === "warnings" ? true : null,
    states,
    platforms,
    agentVersions,
    setQuery,
    setView,
    setStates,
    setPlatforms,
    setAgentVersions,
  };
}
