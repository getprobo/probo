import {
  Environment,
  type FetchFunction,
  Network,
  RecordSource,
  Store,
} from "relay-runtime";
import { GraphQLError } from "graphql";
import type { PropsWithChildren } from "react";
import { RelayEnvironmentProvider } from "react-relay";

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

export function buildEndpoint(path: string): string {
  const host = import.meta.env.VITE_API_URL;

  if (!host) {
    return path;
  }

  const formattedHost =
    host.startsWith("http://") || host.startsWith("https://")
      ? host
      : `https://${host}`;

  const url = new URL(formattedHost);

  if (path) {
    url.pathname = path.startsWith("/") ? path : `/${path}`;
  }

  return url.toString();
}

const hasUnauthenticatedError = (error: GraphQLError) =>
  error.extensions?.code == "UNAUTHENTICATED";

const hasInvalidError = (error: GraphQLError) =>
  error.extensions?.code == "INVALID_REQUEST";

const fetchRelay: FetchFunction = async (
  request,
  variables,
  _,
  uploadables
) => {
  const requestInit: RequestInit = {
    method: "POST",
    credentials: "include",
    headers: {},
  };

  if (uploadables) {
    const formData = new FormData();
    formData.append(
      "operations",
      JSON.stringify({
        operationName: request.name,
        query: request.text,
        variables: variables,
      })
    );

    const uploadableMap: {
      [key: string]: string[];
    } = {};

    Object.keys(uploadables).forEach((key, index) => {
      uploadableMap[index] = [`variables.${key}`];
    });

    formData.append("map", JSON.stringify(uploadableMap));

    Object.keys(uploadables).forEach((key, index) => {
      formData.append(index.toString(), uploadables[key]);
    });

    requestInit.body = formData;
  } else {
    // Extract slug from URL if present for slug-based routing
    const slugMatch = window.location.pathname.match(/^\/trust\/([^/]+)/);
    const slug = slugMatch ? slugMatch[1] : null;

    requestInit.headers = {
      Accept:
        "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
      "Content-Type": "application/json",
      ...(slug ? { "X-Trust-Slug": slug } : {}),
    };

    requestInit.body = JSON.stringify({
      operationName: request.name,
      query: request.text,
      variables,
    });
  }

  // Use relative API path to ensure it goes through the same routing context
  // For /trust/slug/overview, this resolves to /trust/slug/api/trust/v1/graphql
  // For custom domains at /overview, this resolves to /api/trust/v1/graphql
  const response = await fetch(
    buildEndpoint("./api/trust/v1/graphql"),
    requestInit
  );

  if (response.status === 500) {
    throw new InternalServerError();
  }

  const json = await response.json();

  if (json.errors) {
    const errors = json.errors as GraphQLError[];

    if (errors.find(hasUnauthenticatedError)) {
      throw new UnAuthenticatedError();
    }

    const invalidError = errors.find(hasInvalidError);
    if (invalidError) {
      throw new InvalidError(
        invalidError.message,
        invalidError.extensions.field as string ?? "",
        invalidError.extensions.cause as string ?? "",
      );
    }

    throw new Error(`Error fetching GraphQL query '${request.name}'`);
  }

  return json;
};

const source = new RecordSource();
const store = new Store(source, {
  queryCacheExpirationTime: 1 * 60 * 1000,
  gcReleaseBufferSize: 20,
});

export const consoleEnvironment = new Environment({
  configName: "trust",
  network: Network.create(fetchRelay),
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
