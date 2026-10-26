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

import type { GraphQLError } from "graphql";

export type GraphQLSSEPayload = {
  data?: Record<string, unknown> | null;
  errors?: GraphQLError[];
  label?: string;
  path?: ReadonlyArray<string | number> | null;
  hasNext?: boolean;
  extensions?: Record<string, unknown>;
};

export type SSEPush = {
  buffer: string;
  payloads: GraphQLSSEPayload[];
  completed: boolean;
};

// Relay only completes an operation with deferred payloads when the last
// incremental payload carries is_final; gqlgen signals the end with hasNext.
const markRelayFinal = (payload: GraphQLSSEPayload): GraphQLSSEPayload => {
  if (
    payload.hasNext === false
    && payload.path != null
    && payload.label?.includes("$defer$")
  ) {
    return {
      ...payload,
      extensions: { ...payload.extensions, is_final: true },
    };
  }

  return payload;
};

// pushSSE appends a decoded chunk to the pending buffer and returns every
// complete event it now holds, plus the unterminated remainder.
export const pushSSE = (buffer: string, chunk: string): SSEPush => {
  const events = (buffer + chunk).replace(/\r\n?/g, "\n").split("\n\n");
  const rest = events.pop() ?? "";
  const payloads: GraphQLSSEPayload[] = [];

  for (const event of events) {
    let name = "message";
    const data: string[] = [];

    for (const line of event.split("\n")) {
      if (line === "" || line.startsWith(":")) {
        continue;
      }

      const colon = line.indexOf(":");
      const field = colon === -1 ? line : line.slice(0, colon);
      let value = colon === -1 ? "" : line.slice(colon + 1);
      if (value.startsWith(" ")) {
        value = value.slice(1);
      }

      if (field === "event") {
        name = value;
      } else if (field === "data") {
        data.push(value);
      }
    }

    if (name === "complete") {
      return { buffer: "", payloads, completed: true };
    }

    if (name === "next" && data.length > 0) {
      payloads.push(markRelayFinal(JSON.parse(data.join("\n")) as GraphQLSSEPayload));
    }
  }

  return { buffer: rest, payloads, completed: false };
};
