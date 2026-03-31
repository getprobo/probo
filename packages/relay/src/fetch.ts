// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { type FetchFunction } from "relay-runtime";
import {
    InternalServerError,
    UnAuthenticatedError,
    ForbiddenError,
    AssumptionRequiredError,
    NDASignatureRequiredError,
    FullNameRequiredError,
} from "./errors";
import { GraphQLError } from "graphql";

const hasUnauthenticatedError = (error: GraphQLError) =>
    error.extensions?.code == "UNAUTHENTICATED";

const hasFullNameRequiredError = (error: GraphQLError) =>
    error.extensions?.code == "FULL_NAME_REQUIRED";

const hasAssumptionRequiredError = (error: GraphQLError) =>
    error.extensions?.code == "ASSUMPTION_REQUIRED";

const hasNDASignatureRequiredError = (error: GraphQLError) =>
    error.extensions?.code == "NDA_SIGNATURE_REQUIRED";

const hasForbiddenError = (error: GraphQLError) =>
    error.extensions?.code == "FORBIDDEN";

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
                Accept: "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
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

        const json = await response.json();

        if (json.errors) {
            const errors = json.errors as GraphQLError[];

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
                throw new AssumptionRequiredError(assumptionRequiredError.message)
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
