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

export class InternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "InternalServerError";
  }
}

const hasUnauthenticatedError = (error: GraphQLError) =>
  error.extensions?.code == "UNAUTHENTICATED";

const fetchRelay: FetchFunction = async (
  request,
  variables,
  _,
  uploadables,
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
      }),
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
    import.meta.env.VITE_API_URL + "/api/console/v1/query",
    requestInit,
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

    throw new Error(
      `Error fetching GraphQL query '${
        request.name
      }' with variables '${JSON.stringify(variables)}': ${JSON.stringify(
        json.errors,
      )}`,
    );
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

/**
 * Provider for relay with the probo environment
 */
export const RelayProvider = ({ children }: PropsWithChildren) => {
  return (
    <RelayEnvironmentProvider environment={relayEnvironment}>
      {children}
    </RelayEnvironmentProvider>
  );
};
