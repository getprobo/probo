// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { ForbiddenError, InternalServerError, UnAuthenticatedError } from "@probo/relay";
import { type GraphQLError } from "graphql";
import { type FetchFunction, type GraphQLResponse } from "relay-runtime";

// A GraphQL error is "request-level" when it has no `path` — it applies to the
// whole operation (auth, malformed request, transport) rather than a single
// field. Field-level errors carry a `path` and are left in the response so
// Relay can attribute them to the reading field; combined with
// `@throwOnFieldError` on a query/fragment, they surface at the nearest error
// boundary around the component that reads them, instead of collapsing the whole
// operation. See contrib/claude/error-handling.md.
const isRequestLevel = (error: GraphQLError) =>
  error.path === undefined || error.path === null || error.path.length === 0;

// The portal fetch only throws for request-level failures. Everything else
// (field-level errors) flows through to Relay untouched.
export const makeFetchQuery = (endpoint: string): FetchFunction => {
  return async (request, variables, _, uploadables) => {
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
      const uploadableKeys = Object.keys(uploadables);

      uploadableKeys.forEach((key) => {
        uploadableMap[key] = [`variables.${key}`];
      });

      formData.append("map", JSON.stringify(uploadableMap));

      uploadableKeys.forEach((key) => {
        formData.append(key, uploadables[key]);
      });

      requestInit.body = formData;
    } else {
      requestInit.headers = {
        "Accept": "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
        "Content-Type": "application/json",
      };

      requestInit.body = JSON.stringify({
        operationName: request.name,
        query: request.text,
        variables,
      });
    }

    const response = await fetch(endpoint, requestInit);

    if (response.status === 500) {
      throw new InternalServerError();
    }

    const json = (await response.json()) as GraphQLResponse & {
      errors?: GraphQLError[];
    };

    if (json.errors) {
      // An unauthenticated session is always a global concern: the backend
      // attaches the resolver path even to auth errors, so we must scan every
      // error (not just request-level ones) and redirect regardless of where it
      // surfaced.
      const unauthenticated = json.errors.find(
        error => error.extensions?.code === "UNAUTHENTICATED",
      );
      if (unauthenticated) {
        throw new UnAuthenticatedError(unauthenticated.message);
      }

      // Everything else is only thrown here when it is request-level (no path) —
      // a whole-operation failure. Field-level errors (including a FORBIDDEN on
      // a single field/section) are left in the response so Relay surfaces them
      // at the reading component via @throwOnFieldError, containing the failure
      // to that component's boundary instead of the whole page.
      const requestErrors = json.errors.filter(isRequestLevel);

      const forbidden = requestErrors.find(
        error => error.extensions?.code === "FORBIDDEN",
      );
      if (forbidden) {
        throw new ForbiddenError(forbidden.message);
      }

      const requestError = requestErrors[0];
      if (requestError) {
        throw new Error(requestError.message);
      }
    }

    return json;
  };
};
