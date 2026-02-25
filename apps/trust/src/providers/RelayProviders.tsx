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

export class TrustUnAuthenticatedError extends Error {
  constructor() {
    super("UNAUTHENTICATED");
    this.name = "TrustUnAuthenticatedError";
  }
}

export class TrustInvalidError extends Error {
  field?: string;
  cause?: string;

  constructor(message?: string, field?: string, cause?: string) {
    super(message || "INVALID");
    this.name = "TrustInvalidError";
    this.field = field;
    this.cause = cause;
  }
}

export class TrustInternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "TrustInternalServerError";
  }
}

// Legacy exports for backward compatibility
export class UnAuthenticatedError extends TrustUnAuthenticatedError {}
export class InvalidError extends TrustInvalidError {}
export class InternalServerError extends TrustInternalServerError {}

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

export const trustEnvironment = new Environment({
  configName: "trust",
  network: Network.create(makeFetchQuery(buildEndpoint())),
  store,
});

/**
 * Provider for relay with the trust environment
 */
export function TrustRelayProvider({ children }: PropsWithChildren) {
  return (
    <RelayEnvironmentProvider environment={trustEnvironment}>
      {children}
    </RelayEnvironmentProvider>
  );
}

// Legacy export for backward compatibility
export function RelayProvider({ children }: PropsWithChildren) {
  return <TrustRelayProvider>{children}</TrustRelayProvider>;
}
