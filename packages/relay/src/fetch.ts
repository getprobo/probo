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
import { type FetchFunction, type GraphQLResponse } from "relay-runtime";

import {
  AssumptionRequiredError,
  ForbiddenError,
  FullNameRequiredError,
  InternalServerError,
  MembershipRequiredError,
  NDASignatureRequiredError,
  UnAuthenticatedError,
} from "./errors";

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

const isRequestLevel = (error: GraphQLError) =>
  error.path === undefined || error.path === null || error.path.length === 0;

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

    const json = (await response.json()) as ErrorResponse;

    if (json.errors) {
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

      // Field-level errors stay in the response so Relay can attach them to
      // the reading field. Only a request-level failure (no path) throws.
      const requestErrors = errors.filter(isRequestLevel);

      const forbiddenError = requestErrors.find(hasForbiddenError);
      if (forbiddenError) {
        throw new ForbiddenError(forbiddenError.message);
      }

      const requestError = requestErrors[0];
      if (requestError) {
        throw new Error(requestError.message);
      }
    }

    return json;
  };
};
