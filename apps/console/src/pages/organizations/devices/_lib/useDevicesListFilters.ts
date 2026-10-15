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

import {
  devicePlatformFilter,
  type DevicePlatformName,
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

function deviceStateFilter(value: string | null): DeviceStateFilter | null {
  return deviceStates.find(state => state === value) ?? null;
}

function textFilter(value: string | null): string | null {
  if (value == null || value.length === 0) {
    return null;
  }
  return value;
}

export function useDevicesListFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const view: DeviceListView = searchParams.get("view") === "warnings" ? "warnings" : "all";
  const state = deviceStateFilter(searchParams.get("state"));
  const platform = devicePlatformFilter(searchParams.get("platform"));
  const agentVersion = textFilter(searchParams.get("agentVersion"));

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

  const setState = useCallback((value: DeviceStateFilter | null) => {
    setParam("state", value);
  }, [setParam]);

  const setPlatform = useCallback((value: DevicePlatformName | null) => {
    setParam("platform", value);
  }, [setParam]);

  const setAgentVersion = useCallback((value: string | null) => {
    setParam("agentVersion", textFilter(value));
  }, [setParam]);

  return {
    query,
    ownerQuery: devicesListOwnerQuery(query),
    view,
    hasPostureWarning: view === "warnings" ? true : null,
    state,
    platform,
    agentVersion,
    setQuery,
    setView,
    setState,
    setPlatform,
    setAgentVersion,
  };
}
