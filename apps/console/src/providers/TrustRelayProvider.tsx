import {
  Environment,
  type FetchFunction,
  Network,
  RecordSource,
  Store,
} from "relay-runtime";

import type { PropsWithChildren } from "react";
import { RelayEnvironmentProvider } from "react-relay";
import { createContext, useContext, useState } from "react";
import { buildEndpoint } from "./RelayProviders";

export class TrustCenterError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "TrustCenterError";
  }
}

type TrustAuthContextType = {
  isAuthenticated: boolean;
  setAuthenticated: (auth: boolean) => void;
};

const TrustAuthContext = createContext<TrustAuthContextType | null>(null);

export function useTrustAuth() {
  const context = useContext(TrustAuthContext);
  if (!context) {
    throw new Error('useTrustAuth must be used within a TrustRelayProvider');
  }
  return context;
}

let trustEnvironment: Environment | null = null;
let authSetter: ((auth: boolean) => void) | null = null;

function getTrustEnvironment(setAuthenticated: (auth: boolean) => void): Environment {
  if (!trustEnvironment) {
    authSetter = setAuthenticated;

    const trustSource = new RecordSource();
    const trustStore = new Store(trustSource, {
      queryCacheExpirationTime: 5 * 60 * 1000, // 5 minutes
      gcReleaseBufferSize: 10,
    });

    trustEnvironment = new Environment({
      network: Network.create(createFetchTrustRelay()),
      store: trustStore,
    });
  } else {
    authSetter = setAuthenticated;
  }

  return trustEnvironment;
}

const createFetchTrustRelay = (): FetchFunction => async (request, variables) => {
  const requestInit: RequestInit = {
    method: "POST",
    headers: {
      Accept:
        "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({
      operationName: request.name,
      query: request.text,
      variables,
    }),
  };

  const response = await fetch(
    buildEndpoint("/api/trust/v1/graphql"),
    requestInit
  );

  if (response.status === 500) {
    throw new TrustCenterError("Internal server error");
  }

  const json = await response.json();

  if (json.errors?.length > 0) {
    const hasAccessDeniedErrors = json.errors.some((error: any) =>
      error.message.toLowerCase().includes("access denied") ||
      error.message.toLowerCase().includes("unauthorized") ||
      error.extensions?.code === "UNAUTHENTICATED"
    );

    if (hasAccessDeniedErrors) {
      if (authSetter) {
        authSetter(false);
      }
    } else {
      throw new TrustCenterError(
        `Error fetching GraphQL query '${
          request.name
        }' with variables '${JSON.stringify(variables)}': ${JSON.stringify(
          json.errors
        )}`
      );
    }
  }

  return json;
};

export function TrustRelayProvider({ children }: PropsWithChildren) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const environment = getTrustEnvironment(setIsAuthenticated);

  const authContextValue: TrustAuthContextType = {
    isAuthenticated,
    setAuthenticated: setIsAuthenticated,
  };

  return (
    <TrustAuthContext.Provider value={authContextValue}>
      <RelayEnvironmentProvider environment={environment}>
        {children}
      </RelayEnvironmentProvider>
    </TrustAuthContext.Provider>
  );
}
