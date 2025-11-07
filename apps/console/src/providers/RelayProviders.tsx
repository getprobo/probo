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
  constructor(message?: string) {
    super(message || "UNAUTHENTICATED");
    this.name = "UnAuthenticatedError";
  }
}

export class InternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "InternalServerError";
  }
}

export class AuthenticationRequiredError extends Error {
  public redirectUrl: string;
  public requiresSaml: boolean;
  public organizationId: string;
  public samlConfigId?: string;

  constructor(extensions: {
    redirectUrl: string;
    requiresSaml: boolean;
    organizationId: string;
    samlConfigId?: string;
  }) {
    super("AUTHENTICATION_REQUIRED");
    this.name = "AuthenticationRequiredError";
    this.redirectUrl = extensions.redirectUrl;
    this.requiresSaml = extensions.requiresSaml;
    this.organizationId = extensions.organizationId;
    this.samlConfigId = extensions.samlConfigId;
  }
}

export class UnauthorizedError extends Error {
  constructor(message?: string) {
    super(message || "UNAUTHORIZED");
    this.name = "UnauthorizedError";
  }
}

export class ForbiddenError extends Error {
  constructor(message?: string) {
    super(message || "FORBIDDEN");
    this.name = "ForbiddenError";
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

const hasAuthenticationRequiredError = (error: GraphQLError) =>
  error.extensions?.code == "AUTHENTICATION_REQUIRED";

const hasUnauthorizedError = (error: GraphQLError) =>
  error.extensions?.code == "UNAUTHORIZED";

const hasForbiddenError = (error: GraphQLError) =>
  error.extensions?.code == "FORBIDDEN";

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
    requestInit.headers = {
      Accept:
        "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
      "Content-Type": "application/json",
    };

    requestInit.body = JSON.stringify({
      operationName: request.name,
      query: request.text,
      variables,
    });
  }

  const response = await fetch(
    buildEndpoint("/api/console/v1/query"),
    requestInit
  );

  if (response.status === 500) {
    throw new InternalServerError();
  }

  const json = await response.json();

  if (json.errors) {
    const errors = json.errors as GraphQLError[];

    const unauthenticatedError = errors.find(hasUnauthenticatedError);
    if (unauthenticatedError) {
      throw new UnAuthenticatedError(unauthenticatedError.message);
    }

    const authRequiredError = errors.find(hasAuthenticationRequiredError);
    if (authRequiredError?.extensions) {
      const { redirectUrl, requiresSaml, organizationId, samlConfigId } = authRequiredError.extensions;

      throw new AuthenticationRequiredError({
        redirectUrl: redirectUrl as string,
        requiresSaml: requiresSaml as boolean,
        organizationId: organizationId as string,
        samlConfigId: samlConfigId as string | undefined,
      });
    }

    const unauthorizedError = errors.find(hasUnauthorizedError);
    if (unauthorizedError) {
      throw new UnauthorizedError(unauthorizedError.message);
    }

    const forbiddenError = errors.find(hasForbiddenError);
    if (forbiddenError) {
      throw new ForbiddenError(forbiddenError.message);
    }
  }

  return json;
};

const source = new RecordSource();
const store = new Store(source, {
  queryCacheExpirationTime: 1 * 60 * 1000,
  gcReleaseBufferSize: 20,
});

export const relayEnvironment = new Environment({
  network: Network.create(fetchRelay),
  store,
});

export const clearRelayStore = () => {
  const source = relayEnvironment.getStore().getSource();
  if (source instanceof Map) {
    source.clear();
  }
};

/**
 * Provider for relay with the probo environment
 */
export function RelayProvider({ children }: PropsWithChildren) {
  return (
    <RelayEnvironmentProvider environment={relayEnvironment}>
      {children}
    </RelayEnvironmentProvider>
  );
}
