// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { GraphQLError } from "@probo/helpers";

function graphqlCodes(error: unknown): string[] {
  if (typeof error !== "object" || error == null) {
    return [];
  }

  const gqlError = error as GraphQLError;
  const codes: string[] = [];
  if (gqlError.extensions?.code != null) {
    codes.push(gqlError.extensions.code);
  }
  for (const nested of gqlError.source?.errors ?? []) {
    if (nested.extensions?.code != null) {
      codes.push(nested.extensions.code);
    }
  }
  return codes;
}

// Bind preview maps used/expired/missing tokens to GraphQL INVALID. Other
// failures (INTERNAL, network) must not reuse the invalid-token copy.
export function isInvalidBindPreviewError(error: unknown): boolean {
  return graphqlCodes(error).includes("INVALID");
}
