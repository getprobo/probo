import { makeFetchQuery } from "@probo/relay";
import type { PropsWithChildren } from "react";
import { RelayEnvironmentProvider } from "react-relay";
import {
  Environment,
  Network,
  RecordSource,
  Store,
} from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

export class UnAuthenticatedError extends Error {
  constructor() {
    super("UNAUTHENTICATED");
    this.name = "UnAuthenticatedError";
  }
}

export class InvalidError extends Error {
  field?: string;
  cause?: string;

  constructor(message?: string, field?: string, cause?: string) {
    super(message || "INVALID");
    this.name = "InvalidError";
    this.field = field;
    this.cause = cause;
  }
}

export class InternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "InternalServerError";
  }
}

export function buildEndpoint(): string {
  let host = import.meta.env.VITE_API_URL;

  if (!host) {
    host = window.location.origin;
  }

  const formattedHost
    = host.startsWith("http://") || host.startsWith("https://")
      ? host
      : `https://${host}`;

  const url = new URL(formattedHost);

  const prefix = getPathPrefix();
  let path: string;
  if (prefix) {
    path = `${prefix}/api/trust/v1/graphql`;
  } else {
    path = `/api/trust/v1/graphql`;
  }

  url.pathname = path;

  return url.toString();
}

const source = new RecordSource();
const store = new Store(source, {
  queryCacheExpirationTime: 1 * 60 * 1000,
  gcReleaseBufferSize: 20,
});

export const consoleEnvironment = new Environment({
  configName: "trust",
  network: Network.create(makeFetchQuery(buildEndpoint())),
  store,
});

/**
 * Provider for relay with the probo environment
 */
export function RelayProvider({ children }: PropsWithChildren) {
  return (
    <RelayEnvironmentProvider environment={consoleEnvironment}>
      {children}
    </RelayEnvironmentProvider>
  );
}
