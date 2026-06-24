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
  NDASignatureRequiredError,
  UnAuthenticatedError,
} from "./errors";

const hasUnauthenticatedError = (error: GraphQLError) =>
  error.extensions?.code === "UNAUTHENTICATED";

const hasFullNameRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "FULL_NAME_REQUIRED";

const hasAssumptionRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "ASSUMPTION_REQUIRED";

const hasNDASignatureRequiredError = (error: GraphQLError) =>
  error.extensions?.code === "NDA_SIGNATURE_REQUIRED";

const hasForbiddenError = (error: GraphQLError) =>
  error.extensions?.code === "FORBIDDEN";

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

      const forbiddenError = errors.find(hasForbiddenError);
      if (forbiddenError) {
        throw new ForbiddenError(forbiddenError.message);
      }
    }

    return json;
  };
};
