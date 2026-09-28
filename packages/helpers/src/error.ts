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

export interface GraphQLError {
  message?: string;
  extensions?: {
    code?: string;
    field?: string;
    cause?: string;
  };
  source?: {
    errors?: Array<{ message: string; extensions?: { code?: string; field?: string; cause?: string } }>;
  };
}

function graphqlErrorExtension(error: unknown, key: "field" | "cause"): string | undefined {
  if (error == null || typeof error !== "object" || !("extensions" in error)) {
    return undefined;
  }
  const extensions = error.extensions;
  if (extensions == null || typeof extensions !== "object") {
    return undefined;
  }
  const value: unknown = Object.entries(extensions).find(([name]) => name === key)?.[1];
  return typeof value === "string" ? value : undefined;
}

export function graphqlErrorField(error: unknown): string | undefined {
  return graphqlErrorExtension(error, "field");
}

function graphqlErrorMessage(error: unknown): string | undefined {
  if (error == null || typeof error !== "object" || !("message" in error)) {
    return undefined;
  }
  return typeof error.message === "string" ? error.message : undefined;
}

function nestedGraphqlErrors(error: unknown): unknown[] {
  if (error != null && typeof error === "object" && "source" in error) {
    const source = error.source;
    if (
      source != null
      && typeof source === "object"
      && "errors" in source
      && Array.isArray(source.errors)
    ) {
      return source.errors;
    }
  }
  return [error];
}

export type FieldRejection = {
  message: string;
  cause?: string;
};

export function toFieldRejections(error: unknown): Record<string, FieldRejection> | undefined {
  const items = Array.isArray(error) ? error : [error];
  const rejections: Record<string, FieldRejection> = {};

  for (const item of items.flatMap(nestedGraphqlErrors)) {
    const field = graphqlErrorField(item);
    const message = graphqlErrorMessage(item);
    if (field != null && message != null) {
      rejections[field] = { message, cause: graphqlErrorExtension(item, "cause") };
    }
  }

  return Object.keys(rejections).length > 0 ? rejections : undefined;
}

export function toFieldErrors(error: unknown): Record<string, string> | undefined {
  const rejections = toFieldRejections(error);
  if (rejections == null) {
    return undefined;
  }

  return Object.fromEntries(
    Object.entries(rejections).map(([field, { message }]) => [field, message]),
  );
}

export function formatError(title: string, error: GraphQLError | GraphQLError[]): string {
  const messages: string[] = [];

  if (Array.isArray(error)) {
    messages.push(...error.map((e) => e.message).filter(Boolean) as string[]);
  } else if (error.source?.errors && Array.isArray(error.source.errors)) {
    messages.push(...error.source.errors.map((e) => e.message).filter(Boolean));
  } else if (error.message) {
    messages.push(error.message);
  }

  if (messages.length === 0) {
    return title;
  }

  const errorList = messages.join(", ");

  return `${title}: ${errorList}${errorList.endsWith('.') ? '' : '.'}`;
}
