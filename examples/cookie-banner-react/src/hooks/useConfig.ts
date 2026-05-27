import { useCallback, useSyncExternalStore } from "react";

export interface Config {
  bannerId: string;
  baseUrl: string;
}

const STORAGE_KEY = "probo-example-config";

// Seed defaults from Vite env (loaded from ../../../getprobo.com/.env) so the
// example is usable without any manual setup when those vars are present.
const defaultConfig: Config = {
  bannerId: import.meta.env.PUBLIC_COOKIE_BANNER_ID ?? "",
  baseUrl: import.meta.env.PUBLIC_COOKIE_BANNER_API_BASE_URL ?? "",
};

function getSnapshot(): Config {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw) as Config;
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
