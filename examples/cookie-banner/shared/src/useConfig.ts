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

import { useCallback, useSyncExternalStore } from "react";
import { bundledConfig } from "./bundledConfig";

export interface Config {
  bannerId: string;
  baseUrl: string;
  gcmEnabled: boolean;
}

const STORAGE_KEY = "probo-example-config";

const defaultConfig: Config = {
  bannerId: bundledConfig.bannerId,
  baseUrl: bundledConfig.baseUrl,
  gcmEnabled: bundledConfig.gcmEnabled,
};

function getSnapshot(): Config {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return { ...defaultConfig, ...(JSON.parse(raw) as Partial<Config>) };
  } catch {
    // ignore
  }
  return defaultConfig;
}

let cached = getSnapshot();
const listeners = new Set<() => void>();

function subscribe(cb: () => void): () => void {
  listeners.add(cb);
  return () => listeners.delete(cb);
}

function snapshot(): Config {
  return cached;
}

function persist(next: Config): void {
  cached = next;
  localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  for (const cb of listeners) cb();
}

export function useConfig(): [Config, (next: Config) => void] {
  const config = useSyncExternalStore(subscribe, snapshot);
  const setConfig = useCallback((next: Config) => persist(next), []);
  return [config, setConfig];
}
