// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { GraphQLError } from "graphql";
import {
  type FetchFunction,
  type GraphQLResponse,
  Observable,
  type RequestParameters,
} from "relay-runtime";

import {
  AssumptionRequiredError,
  ForbiddenError,
  FullNameRequiredError,
  InternalServerError,
  MembershipRequiredError,
  NDASignatureRequiredError,
  UnAuthenticatedError,
} from "./errors";
import { pushSSE } from "./sse";

type ErrorResponse = GraphQLResponse & { errors?: GraphQLError[] };

const hasCode = (code: string) => (error: GraphQLError) =>
  error.extensions?.code === code;

// Session gates attach a resolver path even when they apply to the whole
// request. Scan every error so the route boundary can still redirect.
const hasUnauthenticatedError = hasCode("UNAUTHENTICATED");
const hasFullNameRequiredError = hasCode("FULL_NAME_REQUIRED");
const hasAssumptionRequiredError = hasCode("ASSUMPTION_REQUIRED");
const hasNDASignatureRequiredError = hasCode("NDA_SIGNATURE_REQUIRED");
const hasMembershipRequiredError = hasCode("MEMBERSHIP_REQUIRED");
const hasForbiddenError = hasCode("FORBIDDEN");

const isPathless = (error: GraphQLError) =>
  error.path === undefined || error.path === null || error.path.length === 0;

// Pathless errors apply to the whole operation. A mutation is a single root
// field, so a path of length 1 is the whole mutation. Query/subscription
// field errors (including root @catch fields) stay in the response.
const isRequestLevel = (error: GraphQLError, operationKind: string) =>
  isPathless(error)
  || (operationKind === "mutation" && error.path?.length === 1);

const checkResponse = (json: ErrorResponse, operationKind: string) => {
  if (!json.errors) {
    return json;
  }

  const errors = json.errors;

  const unauthenticatedError = errors.find(hasUnauthenticatedError);
  if (unauthenticatedError) {
    throw new UnAuthenticatedError(unauthenticatedError.message);
  }

  const fullNameRequiredError = errors.find(hasFullNameRequiredError);
  if (fullNameRequiredError) {
    throw new FullNameRequiredError(fullNameRequiredError.message);
  }

  const assumptionRequiredError = errors.find(hasAssumptionRequiredError);
  if (assumptionRequiredError) {
    throw new AssumptionRequiredError(assumptionRequiredError.message);
  }

  const ndaSignatureRequiredError = errors.find(hasNDASignatureRequiredError);
  if (ndaSignatureRequiredError) {
    throw new NDASignatureRequiredError(ndaSignatureRequiredError.message);
  }

  const membershipRequiredError = errors.find(hasMembershipRequiredError);
  if (membershipRequiredError) {
    throw new MembershipRequiredError(membershipRequiredError.message);
  }

  // Nested field errors stay in the response so Relay can attach them
  // to the reading field. Pathless errors always throw. Mutation-root
  // errors (path length 1) also throw so ForbiddenError still maps.
  const requestErrors = errors.filter(error =>
    isRequestLevel(error, operationKind),
  );

  const forbiddenError = requestErrors.find(hasForbiddenError);
  if (forbiddenError) {
    throw new ForbiddenError(forbiddenError.message);
  }

  const requestError = requestErrors[0];
  if (requestError) {
    throw new Error(requestError.message);
  }

  return json;
};

// Only deferred operations ask for an event stream: the SSE transport sends
// headers before resolvers run, so cookies set by a mutation would be lost.
const isDeferred = (request: RequestParameters) =>
  request.operationKind !== "mutation"
  && (request.text?.includes("@defer") ?? false);

const fetchMultipart = async (
  endpoint: string,
  request: RequestParameters,
  variables: Record<string, unknown>,
  uploadables: Record<string, File | Blob>,
) => {
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
  const uploadableKeys = Object.keys(uploadables);

  uploadableKeys.forEach((key) => {
    uploadableMap[key] = [`variables.${key}`];
  });

  formData.append("map", JSON.stringify(uploadableMap));

  uploadableKeys.forEach((key) => {
    formData.append(key, uploadables[key]);
  });

  const response = await fetch(endpoint, {
    method: "POST",
    credentials: "include",
    body: formData,
  });

  if (response.status === 500) {
    throw new InternalServerError();
  }

  return checkResponse(
    (await response.json()) as ErrorResponse,
    request.operationKind,
  );
};

const fetchJSON = (
  endpoint: string,
  request: RequestParameters,
  variables: Record<string, unknown>,
) =>
  Observable.create<GraphQLResponse>((sink) => {
    const controller = new AbortController();
    const deferred = isDeferred(request);

    const run = async () => {
      const response = await fetch(endpoint, {
        method: "POST",
        credentials: "include",
        signal: controller.signal,
        headers: {
          "Accept": deferred
            ? "text/event-stream, application/graphql-response+json; charset=utf-8, application/json; charset=utf-8"
            : "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          operationName: request.name,
          query: request.text,
          variables,
        }),
      });

      if (response.status === 500) {
        throw new InternalServerError();
      }

      const contentType = response.headers.get("Content-Type") ?? "";
      if (!contentType.includes("text/event-stream") || !response.body) {
        sink.next(checkResponse(
          (await response.json()) as ErrorResponse,
          request.operationKind,
        ));
        sink.complete();
        return;
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      try {
        for (;;) {
          const { done, value } = await reader.read();
          if (done) {
            throw new Error("GraphQL event stream ended before completion");
          }

          const push = pushSSE(buffer, decoder.decode(value, { stream: true }));
          buffer = push.buffer;

          for (const payload of push.payloads) {
            sink.next(checkResponse(
              payload as ErrorResponse,
              request.operationKind,
            ));
          }

          if (push.completed) {
            sink.complete();
            return;
          }
        }
      } finally {
        reader.cancel().catch(() => {});
      }
    };

    run().catch((error: unknown) => {
      if (controller.signal.aborted) {
        return;
      }

      sink.error(error instanceof Error ? error : new Error(String(error)));
    });

    return () => controller.abort();
  });

export const makeFetchQuery = (endpoint: string): FetchFunction => {
  return (request, variables, _, uploadables) => {
    if (uploadables) {
      return fetchMultipart(endpoint, request, variables, uploadables);
    }

    return fetchJSON(endpoint, request, variables);
  };
};
