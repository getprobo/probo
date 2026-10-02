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
  Observable,
  type FetchFunction,
  type GraphQLResponse,
  type RequestParameters,
  type UploadableMap,
  type Variables,
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
import { pushSSE, type GraphQLSSEPayload } from "./sse";

const hasUnauthenticatedError = (error: GraphQLError) =>
  error.extensions?.code === "UNAUTHENTICATED";

const hasFullNameRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "FULL_NAME_REQUIRED";

const hasAssumptionRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "ASSUMPTION_REQUIRED";

const hasNDASignatureRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "NDA_SIGNATURE_REQUIRED";

const hasMembershipRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "MEMBERSHIP_REQUIRED";

const hasForbiddenError = (error: GraphQLError) =>
  error.extensions?.code === "FORBIDDEN";

function classifyGraphQLErrors(errors: readonly GraphQLError[] | undefined) {
  if (errors == null) {
    return;
  }

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

  const forbiddenError = errors.find(hasForbiddenError);
  if (forbiddenError) {
    throw new ForbiddenError(forbiddenError.message);
  }
}

function isAbortError(error: unknown) {
  return error instanceof Error && error.name === "AbortError";
}

async function readJSONResponse(response: Response): Promise<GraphQLResponse> {
  const json = (await response.json()) as GraphQLResponse & {
    errors?: GraphQLError[];
  };
  classifyGraphQLErrors(json.errors);
  return json;
}

async function fetchUpload(
  endpoint: string,
  request: RequestParameters,
  variables: Variables,
  uploadables: UploadableMap,
): Promise<GraphQLResponse> {
  const formData = new FormData();
  formData.append(
    "operations",
    JSON.stringify({
      operationName: request.name,
      query: request.text,
      variables,
    }),
  );

  const uploadableMap: { [key: string]: string[] } = {};
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

  return readJSONResponse(response);
}

async function readEventStream(
  response: Response,
  publish: (payload: GraphQLSSEPayload) => void,
): Promise<void> {
  if (response.body == null) {
    throw new InternalServerError();
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let completed = false;

  try {
    while (!completed) {
      const { done, value } = await reader.read();
      const chunk = value == null
        ? decoder.decode()
        : decoder.decode(value, { stream: !done });
      const pushed = pushSSE(buffer, chunk);
      buffer = pushed.buffer;
      completed = pushed.completed;
      for (const payload of pushed.payloads) {
        classifyGraphQLErrors(payload.errors);
        publish(payload);
      }
      if (done) {
        break;
      }
    }
  } finally {
    await reader.cancel().catch(() => undefined);
  }

  if (!completed) {
    throw new Error("graphql event stream ended before complete");
  }
}

export const makeFetchQuery = (endpoint: string): FetchFunction => {
  return (request, variables, _, uploadables) => {
    if (uploadables) {
      return fetchUpload(endpoint, request, variables, uploadables);
    }

    return Observable.create((sink) => {
      const controller = new AbortController();

      void (async () => {
        try {
          const response = await fetch(endpoint, {
            method: "POST",
            credentials: "include",
            signal: controller.signal,
            headers: {
              "Accept": [
            "text/event-stream",
            "application/graphql-response+json; charset=utf-8",
            "application/json; charset=utf-8",
          ].join(", "),
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              operationName: request.name,
              query: request.text,
              variables,
            }),
          });

          if (response.status === 500) {
            sink.error(new InternalServerError());
            return;
          }

          const contentType = response.headers.get("content-type") ?? "";
          if (!contentType.includes("text/event-stream")) {
            sink.next(await readJSONResponse(response));
            sink.complete();
            return;
          }

          await readEventStream(response, (payload) => {
            sink.next(payload as GraphQLResponse);
          });
          sink.complete();
        } catch (error) {
          if (isAbortError(error) || sink.closed) {
            return;
          }
          sink.error(error instanceof Error ? error : new Error("graphql request failed"));
        }
      })();

      return () => {
        controller.abort();
      };
    });
  };
};
