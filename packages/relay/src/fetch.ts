import { type FetchFunction } from "relay-runtime";
import {
    InternalServerError,
    UnAuthenticatedError,
    AuthenticationRequiredError,
    UnauthorizedError,
    ForbiddenError,
} from "./errors";
import { GraphQLError } from "graphql";

const hasUnauthenticatedError = (error: GraphQLError) =>
    error.extensions?.code == "UNAUTHENTICATED";

const hasAuthenticationRequiredError = (error: GraphQLError) =>
    error.extensions?.code == "AUTHENTICATION_REQUIRED";

const hasUnauthorizedError = (error: GraphQLError) =>
    error.extensions?.code == "UNAUTHORIZED";

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

            const authRequiredError = errors.find(hasAuthenticationRequiredError);
            if (authRequiredError?.extensions) {
                const { redirectUrl, requiresSaml, organizationId, samlConfigId } =
                    authRequiredError.extensions;

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
};
